package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

func logRequest(prefix string, r *http.Request) {
	log.Printf("%s %s %s %s %s", prefix, r.Host, r.Method, r.URL, r.Proto)
}

type Proxy struct {
	ProxyRules []Rule
}

type Rule struct {
	RequestPathRegex string
	ForwardTo        string
}

var (
	hostFlag     string // hostname to listen to
	portFlag     int    // server port
	confFileFlag string // location of configuration file
)

func init() {
	flag.StringVar(&hostFlag, "host", "localhost", "Reverse proxy host")
	flag.IntVar(&portFlag, "port", 8080, "Reverse proxy port")
	flag.StringVar(&confFileFlag, "conf", "", `Location of the configuration file
	Example configuration below. Create a json file and pass the file location
	{
	        "ProxyRules": [
	                {
	                        "RequestPathRegex": "^/services/SimpleStockQuoteService",
	                        "ForwardTo": "http://localhost:9000"
	                },
	                {
	                        "RequestPathRegex": "^/services/FastStockQuoteService",
	                        "ForwardTo": "http://localhost:9001"
	                }
	        ]
	}
		`)
}

func main() {
	flag.Parse()

	if confFileFlag == "" {
		log.Fatal("Conf file not found. Use -h to find out config file format")
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	log.Print("Reading proxy rules... ")
	cf, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal("Error reading config file", err.Error())
	}

	var config Proxy
	err = json.Unmarshal(cf, &config)
	log.Print("Loaded proxy rules from config.json")

	if len(config.ProxyRules) == 0 {
		log.Fatal("There's no proxy rules, please add rules to config.json")
		panic(1)
	}

	log.Printf("Listening on %s:%d", hostFlag, portFlag)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		var matched = false

		for _, rule := range config.ProxyRules {
			matched, _ = regexp.MatchString(rule.RequestPathRegex, r.RequestURI)

			if matched {
				logRequest("[incoming]", r)

				client := &http.Client{}
				req, _ := http.NewRequest(r.Method, fmt.Sprintf("%s%s", rule.ForwardTo, r.RequestURI), r.Body)

				// Add all HTTP headers in request
				for k, v := range r.Header {
					req.Header.Add(k, v[0])
				}
				logRequest("[upstream]", req)
				resp, err := client.Do(req)

				if err != nil {
					log.Println(err.Error())
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
				} else {
					// Successful response from backend
					w.WriteHeader(http.StatusOK)
					_, _ = io.Copy(w, resp.Body)
				}

				// If matched, don't have to continue
				break
			}
		}

		if !matched {
			log.Println("writing header - not found")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found\n"))
		}

	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", hostFlag, portFlag), nil))
}
