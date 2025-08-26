## Bento testing

Trying to see if 

- it's safe to have different streams modify the same cache key.
- how the `system-window` buffer works

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


```
$ go run main.go 
Starting Bento cache test...
Press Ctrl+C to stop both streams

INFO Receiving inproc messages from ID: f850f06c-491c-463a-992b-a84c12ac2e27  @service=bento label="" path=root.input
INFO Sending inproc messages to ID: 9c0f14f7-d0cc-4c52-841e-4b546d73c09d  @service=bento label="" path=root.output
sending message { "traffic_light": "two",  "created_at": "2025-08-25T22:09:42-07:00", "passengers": 2}
sending message { "traffic_light": "two",  "created_at": "2025-08-25T22:09:42-07:00", "passengers": 7}
sending message { "traffic_light": "two",  "created_at": "2025-08-25T22:09:42-07:00", "passengers": 1}
sending message { "traffic_light": "one",  "created_at": "2025-08-25T22:09:42-07:00", "passengers": 0}
sending message { "traffic_light": "one",  "created_at": "2025-08-25T22:09:42-07:00", "passengers": 8}
-> consuming message {"created_at":"2025-08-26T05:09:50Z","passengers":10,"traffic_light":"two"}
-> consuming message {"created_at":"2025-08-26T05:09:50Z","passengers":8,"traffic_light":"one"}
```