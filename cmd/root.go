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
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/kklipsch/comed_exporter/api"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	applicationLogger = log.WithFields(log.Fields{"name": "comed_exporter"})

	rootCmd = &cobra.Command{
		Use:   "comed_exporter",
		Short: "exposes the comed hourly pricing as prometheus metrics",
		RunE:  start,
	}
)

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
	rootCmd.Flags().String("api", api.Address, "The comed api endpoint")
	rootCmd.Flags().Duration("schedule", time.Minute*5, "How often to query the api")
}

func initConfig() {
	viper.SetEnvPrefix("comed_exporter")
	viper.AutomaticEnv()

	viper.BindPFlag("address", rootCmd.Flags().Lookup("address"))
	viper.BindPFlag("api", rootCmd.Flags().Lookup("api"))
	viper.BindPFlag("schedule", rootCmd.Flags().Lookup("schedule"))
}

func start(cmd *cobra.Command, args []string) error {
	applicationLogger.Infoln("starting up")
	ctx := setSignalCancel(context.Background(), os.Interrupt, os.Kill, syscall.SIGTERM)

	srv := startServer(viper.GetString("address"))

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	var err error
	client.Transport, err = instrumentClient("query", client.Transport)
	if err != nil {
		return err
	}

	go startQuerying(ctx, viper.GetString("api"), viper.GetDuration("schedule"), client)

	applicationLogger.Infoln("started")

	<-ctx.Done()

	shutdownCtx, clean := context.WithTimeout(context.Background(), time.Second*5)
	defer clean()

	err = srv.Shutdown(shutdownCtx)
	if err != nil {
		return fmt.Errorf("error shutting down web server: %v", err)
	}

	applicationLogger.Infoln("done")
	return nil

}

func startQuerying(ctx context.Context, address string, schedule time.Duration, client *http.Client) {
	//do once on startup so you dont have to wait for the ticker
	doQuery(client, address)
	t := time.Tick(schedule)

	for {
		select {
		case <-t:
			doQuery(client, address)
		case <-ctx.Done():
			return
		}
	}
}

func doQuery(client *http.Client, address string) {
	price, err := api.GetLastPrice(client, address)
	if err != nil {
		errorsCount.Inc()
		applicationLogger.WithField("err", err).Errorln("error querying api")
		return
	}

	priceGuage.Set(price.CentsPerKWh)
}

func startServer(address string) *http.Server {
	srv := &http.Server{Addr: address, Handler: endpoint()}
	go func() {
		err := srv.ListenAndServe()
		if err != http.ErrServerClosed {
			applicationLogger.WithFields(log.Fields{"err": err}).Fatalln("failed at serving")
		}
	}()

	return srv
}

func endpoint() http.Handler {
	router := httprouter.New()
	router.Handler("GET", "/metrics", instrumentHandler("metrics", promhttp.Handler()))
	return router
}

func setSignalCancel(ctx context.Context, sig ...os.Signal) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, sig...)

	go func() {
		<-sigChan
		applicationLogger.WithFields(log.Fields{"signal": sig}).Println("received stop signal")
		cancel()
	}()

	return ctx
}
