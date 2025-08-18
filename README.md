## Bento cache test

Trying to see if it's safe to have different streams modify the same cache key.

```
$ go run main.go
Starting Bento cache test...
Press Ctrl+C to stop both streams

2025/08/17 18:47:23 Starting Stream A...
2025/08/17 18:47:23 Starting Stream B...
INFO Input type generate is now active             @service=bento label="" path=root.input
INFO Input type generate is now active             @service=bento label="" path=root.input
INFO Output type stdout is now active              @service=bento label="" path=root.output
INFO Output type stdout is now active              @service=bento label="" path=root.output
{"b":1}
{"a":1}
{"a":2}
{"a":1,"b":1}
{"a":1,"b":2}
{"a":2,"b":1}
{"a":2,"b":2}
{"a":3,"b":1}
{"a":4,"b":1}
{"a":3,"b":2}
{"a":3,"b":3}
{"a":4,"b":2}
{"a":5,"b":2}
{"a":4,"b":3}
```