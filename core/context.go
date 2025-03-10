package core

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/yomorun/yomo/core/frame"
	"github.com/yomorun/yomo/core/metadata"
	"github.com/yomorun/yomo/core/router"
	"golang.org/x/exp/slog"
)

var ctxPool sync.Pool

// Context is context for stream handling.
// Context is generated subsequent to the arrival of a dataStream and retains pertinent information derived from the dataStream. The lifespan of the Context should align with the lifespan of the Stream.
type Context struct {
	// DataStream is the stream used for reading and writing frames.
	DataStream DataStream
	// Frame receives from client.
	Frame *frame.DataFrame
	// FrameMetadata is the merged metadata from the frame.
	FrameMetadata metadata.M
	// Route is the route from handshake.
	Route router.Route
	// mu is used to protect Keys from concurrent read and write operations.
	mu sync.RWMutex
	// Keys stores the key/value pairs in context, It is Lazy initialized.
	Keys map[string]any
	// StreamLogger is stream-level logger.
	StreamLogger *slog.Logger
	// Using Logger to log in stream handler scope, Logger is frame-level logger.
	Logger *slog.Logger
}

// Set is used to store a new key/value pair exclusively for this context.
// It also lazy initializes c.Keys if it was not used previously.
func (c *Context) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Keys == nil {
		c.Keys = make(map[string]any)
	}

	c.Keys[key] = value
}

// Get returns the value for the given key, ie: (value, true).
// Returns (nil, false) if the value does not exist.
func (c *Context) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, ok := c.Keys[key]
	return value, ok
}

var _ context.Context = &Context{}

// Done returns nil (chan which will wait forever) when c.Stream.Context() has no Context.
func (c *Context) Done() <-chan struct{} { return c.DataStream.Context().Done() }

// Deadline returns that there is no deadline (ok==false) when c.Stream has no Context.
func (c *Context) Deadline() (deadline time.Time, ok bool) { return c.DataStream.Context().Deadline() }

// Err returns nil when c.Request has no Context.
func (c *Context) Err() error { return c.DataStream.Context().Err() }

// Value retrieves the value associated with the specified key within the context.
// If no value is found, it returns nil. Subsequent invocations of "Value" with the same key yield identical outcomes.
func (c *Context) Value(key any) any {
	c.mu.Lock()
	if keyAsString, ok := key.(string); ok {
		if val, exists := c.Keys[keyAsString]; exists {
			c.mu.Unlock()
			return val
		}
	}
	c.mu.Unlock()

	// this will not take effect forever.
	return c.DataStream.Context().Value(key)
}

// newContext returns a new YoMo context that implements the standard library `context.Context` interface.
// The YoMo context is used to manage the lifecycle of a stream and provides a way to pass data and metadata between stream processing functions.
// The lifecycle of the context is equal to the lifecycle of the stream that it is associated with.
// The context can be used to manage timeouts, cancellations, and other aspects of stream processing.
func newContext(dataStream DataStream, route router.Route, logger *slog.Logger) (c *Context) {
	v := ctxPool.Get()
	if v == nil {
		c = new(Context)
	} else {
		c = v.(*Context)
	}

	logger = logger.With(
		"stream_id", dataStream.ID(),
		"stream_name", dataStream.Name(),
		"stream_type", dataStream.StreamType().String(),
	)

	c.DataStream = dataStream
	c.Route = route
	c.StreamLogger = logger
	c.Logger = logger

	return
}

// WithFrame sets the current frame of the YoMo context to the given data frame.
// It extracts the metadata from the data frame and sets it as attributes on the context logger.
// It also merges the metadata from the data stream with the metadata from the data frame.
// This allows downstream processing functions to access the metadata from both the data stream and the current data frame.
// If the given frame is not a data frame, it returns an error.
// If there is an error decoding the metadata from the data frame, it returns that error.
// Otherwise, it sets the current frame and frame metadata on the context and returns nil.
func (c *Context) WithFrame(f frame.Frame) error {
	df, ok := f.(*frame.DataFrame)
	if !ok {
		return errors.New("data stream only transmit data frame")
	}

	fmd, err := metadata.Decode(df.Metadata)
	if err != nil {
		return err
	}

	c.Logger = c.StreamLogger.With(MetadataSlogAttr(fmd))

	// merge data stream metadata.
	c.DataStream.Metadata().Range(func(k, v string) bool {
		fmd.Set(k, v)
		return true
	})

	c.Frame = df
	c.FrameMetadata = fmd

	return nil
}

// CloseWithError close dataStream with an error string.
func (c *Context) CloseWithError(errString string) {
	c.Logger.Debug("data stream closed", "error", errString)

	err := c.DataStream.Close()
	if err == nil {
		return
	}
	c.Logger.Error("data stream close failed", "err", err)
}

// Release release the Context, the Context which has been released will not be available.
//
// Warning: do not use any Context api after Release, It maybe cause an error.
// TODO: use a state to ensure safe access and release of the context.
func (c *Context) Release() {
	c.reset()
	ctxPool.Put(c)
}

func (c *Context) reset() {
	c.DataStream = nil
	c.Route = nil
	c.Frame = nil
	c.FrameMetadata = nil
	c.StreamLogger = nil
	c.Logger = nil
	for k := range c.Keys {
		delete(c.Keys, k)
	}
}
