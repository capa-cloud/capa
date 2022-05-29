package main

import (
	"group.rxcloud/capa/pkg/runtime"
	"os"
	"os/signal"
	"syscall"
)

//func newRuntimeApp(startCmd *cli.Command) *cli.App {
//	app := cli.NewApp()
//	app.Name = "Layotto"
//	app.Version = GitVersion
//	app.Compiled = time.Now()
//	app.Copyright = "(c) " + strconv.Itoa(time.Now().Year()) + " Layotto Authors"
//	app.Usage = "A fast and efficient cloud native application runtime based on MOSN."
//	app.Flags = cmdStart.Flags
//
//	// commands
//	app.Commands = []cli.Command{
//		cmdStart,
//	}
//	// action
//	app.Action = func(c *cli.Context) error {
//		if c.NumFlags() == 0 {
//			return cli.ShowAppHelp(c)
//		}
//
//		return startCmd.Action.(func(c *cli.Context) error)(c)
//	}
//
//	return app
//}
//
//func (p *CapaRuntime) ServeHTTP(w http.ResponseWriter, req *http.Request) {
//	// Forward the HTTP request to the destination service.
//	res, duration, err := p.forwardRequest(req)
//
//	// Notify the client if there was an error while forwarding the request.
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusBadGateway)
//		return
//	}
//
//	// If the request was forwarded successfully, write the response back to
//	// the client.
//	p.writeResponse(w, res)
//
//	// Print request and response statistics.
//	p.printStats(req, res, duration)
//}
//
//func (p *CapaRuntime) forwardRequest(req *http.Request) (*http.Response, time.Duration, error) {
//	// Prepare the destination endpoint to forward the request to.
//	proxyUrl := fmt.Sprintf("http://127.0.0.1:%d%s", servicePort, req.RequestURI)
//
//	// Print the original URL and the proxied request URL.
//	fmt.Printf("Original URL: http://%s:%d%s\n", req.Host, servicePort, req.RequestURI)
//	fmt.Printf("CapaRuntime URL: %s\n", proxyUrl)
//
//	// Create an HTTP client and a proxy request based on the original request.
//	httpClient := http.Client{}
//	proxyReq, err := http.NewRequest(req.Method, proxyUrl, req.Body)
//
//	// Capture the duration while making a request to the destination service.
//	start := time.Now()
//	res, err := httpClient.Do(proxyReq)
//	duration := time.Since(start)
//
//	// Return the response, the request duration, and the error.
//	return res, duration, err
//}
//func (p *CapaRuntime) writeResponse(w http.ResponseWriter, res *http.Response) {
//	// Copy all the header values from the response.
//	for name, values := range res.Header {
//		w.Header()[name] = values
//	}
//
//	// Set a special header to notify that the proxy actually serviced the request.
//	w.Header().Set("Server", "amazing-proxy")
//
//	// Set the status code returned by the destination service.
//	w.WriteHeader(res.StatusCode)
//
//	// Copy the contents from the response body.
//	io.Copy(w, res.Body)
//
//	// Finish the request.
//	res.Body.Close()
//}
//func (p *CapaRuntime) printStats(req *http.Request, res *http.Response, duration time.Duration) {
//	fmt.Printf("Request Duration: %v\n", duration)
//	fmt.Printf("Request Size: %d\n", req.ContentLength)
//	fmt.Printf("Response Size: %d\n", res.ContentLength)
//	fmt.Printf("Response Status: %d\n\n", res.StatusCode)
//}

func main() {
	// Listen on the predefined proxy port.
	rt := runtime.NewCapaRuntime(&runtime.CapaRuntimeConfig{
		AppManagement: &runtime.AppConfig{
			AppId: "123",
			Env:   "local",
			Cloud: "local",
		},
		SidecarManagement: &runtime.SidecarConfig{
			APIListenAddresses:       []string{"127.0.0.1"},
			RuntimePort:              8081,
			RuntimeCallbackPort:      8081,
			GracefulShutdownDuration: 10,
		},
	})
	//http.ListenAndServe(fmt.Sprintf(":%d", proxyPort), cr)
	rt.Run(runtime.WithActors(nil))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, os.Interrupt)
	<-stop
	rt.ShutdownWithWait()
}
