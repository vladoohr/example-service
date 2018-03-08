package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/JormungandrK/microservice-tools/gateway"
)

func main() {
	var port = flag.Int("p", 8080, "Listen port.")
	var gwAdminURL = flag.String("gw", "http://kong:8001", "API Gateway admin url")
	var serviceName = flag.String("name", "", "The name of this service")
	var serviceDomain = flag.String("domain", "service.consul", "Internal domain for the server.")
	var path = flag.String("path", "/example", "Path pattern used for routing requests.")
	var skipRegister = flag.Bool("skipgw", false, "Skip Gateway self-registration.")

	flag.Parse()

	log.Printf("Service Configuration:\n\t* Port: %d\n\t* Gateway Admin URL: %s\n\t* Service Name: %s\n\t* Domain: %s\n\t* Path: %s\n", *port, *gwAdminURL, *serviceName, *serviceDomain, *path)

	if serviceName == nil || *serviceName == "" {
		*serviceName = getEnv("SERVICE_NAME", "example-service")
	}

	if !*skipRegister {
		gw := gateway.NewKongGateway(*gwAdminURL, &http.Client{}, &gateway.MicroserviceConfig{
			MicroserviceName: *serviceName,
			MicroservicePort: *port,
			Paths:            []string{*path},
			VirtualHost:      fmt.Sprintf("%s.%s", *serviceName, *serviceDomain),
		})

		if err := gw.SelfRegister(); err != nil {
			log.Fatal("Failed to register on the API Gateway", err.Error())
		}
	} else {
		log.Println("Skipped Gateway registration.")
	}

	http.HandleFunc(*path, func(rw http.ResponseWriter, req *http.Request) {
		start := time.Now().UnixNano()
		switch method := req.Method; method {
		case "GET":
			data, err := json.Marshal(map[string]interface{}{
				"message": "Hello. This is example service.",
			})
			if err != nil {
				writeError(err, rw)
				break
			}
			rw.Header().Add("Content-Type", "application/json")
			rw.WriteHeader(200)
			rw.Write(data)
			break
		case "POST", "PUT", "PATCH":
			data := map[string]interface{}{}
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				writeError(err, rw)
				break
			}
			if body == nil || len(body) == 0 {
				body = []byte("{}")
			}
			if err = json.Unmarshal(body, &data); err != nil {
				writeError(err, rw)
				break
			}

			resp, err := json.Marshal(map[string]interface{}{
				"message": "Hello. This is example service. Here is what you've send:",
				"method":  req.Method,
				"body":    data,
			})
			if err != nil {
				writeError(err, rw)
				break
			}

			rw.Header().Add("Content-Type", "application/json")
			rw.WriteHeader(200)
			rw.Write(resp)
			break
		default:
			fmt.Println("C")
			rw.WriteHeader(405)
			rw.Write([]byte(fmt.Sprintf("Method %s is not allowed.", req.Method)))
			break

		}
		end := time.Now().UnixNano()
		log.Println("Req", req.URL, fmt.Sprintf("handled in %.2fμs.", float64(end-start)/1000.0))
	})
	address := fmt.Sprintf(":%d", *port)
	log.Println("Listening on", address)
	log.Fatal(http.ListenAndServe(address, nil))

}

func writeError(err error, rw http.ResponseWriter) {
	rw.WriteHeader(500)
	rw.Write([]byte(err.Error()))
}

func getEnv(key, defValue string) string {
	val := os.Getenv(key)
	if val == "" {
		val = defValue
	}
	return val
}
