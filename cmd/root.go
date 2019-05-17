// Copyright © 2019 Kasey Klipsch
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "comed_exporter",
	Short: "exposes the comed hourly pricing as prometheus metrics",
	Run:   start,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Flags().String("address", ":9010", "The address to return results on.")
	rootCmd.Flags().String("api", "https://hourlypricing.comed.com/api", "The comed api endpoint")
	rootCmd.Flags().Duration("schedule", time.Minute*5, "How often to query the api")
}

func initConfig() {
	viper.SetEnvPrefix("comed_exporter")
	viper.AutomaticEnv()

	viper.BindPFlag("address", rootCmd.Flags().Lookup("address"))
	viper.BindPFlag("api", rootCmd.Flags().Lookup("api"))
	viper.BindPFlag("schedule", rootCmd.Flags().Lookup("schedule"))
}

func start(cmd *cobra.Command, args []string) {
	fmt.Printf("address: %s\n", viper.GetString("address"))
	fmt.Printf("api: %s\n", viper.GetString("api"))
	fmt.Printf("schedule: %s\n", viper.GetDuration("schedule"))
}
