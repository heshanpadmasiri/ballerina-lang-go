// Copyright (c) 2026, WSO2 LLC. (http://www.wso2.com).
//
// WSO2 LLC. licenses this file to you under the Apache License,
// Version 2.0 (the "License"); you may not use this file except
// in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package main

import (
	"fmt"
	"os"

	"ballerina-lang-go/lsp"

	"github.com/spf13/cobra"
)

var lspCmd = &cobra.Command{
	Use:   "lsp",
	Short: "Start the Ballerina language server",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := profiler.Start(); err != nil {
			return fmt.Errorf("failed to start profiler: %w", err)
		}
		defer func() { _ = profiler.Stop() }()

		server := lsp.NewServer(os.Stdin, os.Stdout)
		return server.Run()
	},
}
