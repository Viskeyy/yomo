/*
Copyright © 2021 Allegro Networks

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yomorun/yomo/cli/serverless/golang"
	"github.com/yomorun/yomo/pkg/file"
	"github.com/yomorun/yomo/pkg/log"
)

var (
	name string
	rx   bool
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a YoMo Stream function",
	Long:  "Initialize a YoMo Stream function",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) >= 1 && args[0] != "" {
			name = args[0]
		}

		if name == "" {
			log.FailureStatusEvent(os.Stdout, "Please input your app name")
			return
		}

		log.PendingStatusEvent(os.Stdout, "Initializing the Stream Function...")
		name = strings.ReplaceAll(name, " ", "_")
		// create app.go
		fname := filepath.Join(name, defaultSFNSourceFile)
		contentTmpl := golang.InitTmpl
		if rx {
			contentTmpl = golang.InitRxTmpl
		}
		if err := file.PutContents(fname, contentTmpl); err != nil {
			log.FailureStatusEvent(os.Stdout, "Write stream function into app.go file failure with the error: %v", err)
			return
		}

		// create .env
		fname = filepath.Join(name, ".env")
		if err := file.PutContents(fname, []byte(fmt.Sprintf("YOMO_SFN_NAME=%s\n", name))); err != nil {
			log.FailureStatusEvent(os.Stdout, "Write stream function .env file failure with the error: %v", err)
			return
		}

		log.SuccessStatusEvent(os.Stdout, "Congratulations! You have initialized the stream function successfully.")
		log.InfoStatusEvent(os.Stdout, "You can enjoy the YoMo Stream Function via the command: ")
		log.InfoStatusEvent(os.Stdout, "\tStep 1: cd %s && yomo build", name)
		log.InfoStatusEvent(os.Stdout, "\tStep 2: yomo run sfn.wasm")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVarP(&name, "name", "n", "", "The name of Stream Function")
	initCmd.Flags().BoolVarP(&rx, "rx", "r", false, "Use RxStream Handler")
}
