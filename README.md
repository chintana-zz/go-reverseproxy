# go-reverseproxy

Simple reverse proxy impl in Go.

## Running

<code>
$ go get github.com/chintana/go-reverseproxy
</code>

Create a config file with proxy rules. <code>go-reverseproxy -h</code> for more info.

Example config,
<pre><code>
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
</code></pre>

Then run with <code>$ go-reverseproxy -conf config.json</code>
