#About

The package is for daily log.

#Dependencies

1. [github.com/pkg/errors](https://github.com/pkg/errors)
   ```
   go get github.com/pkg/errors
   ```

#Benchmark

##Test methods out a line to the log

The result with `o.outC <- fmt.Sprint(args...)`. It is
```
50000	     28158 ns/op	    1862 B/op	       9 allocs/op
PASS
```
with one args and is
```
5 args
50000	     31835 ns/op	    2108 B/op	      13 allocs/op
PASS

10 args
30000	     41822 ns/op	    2783 B/op	      20 allocs/op
PASS
```
with multiple args.


The result with `o.outC <- o.outC <- string(a) // a := make([]byte, 0, 128)`. It is
```
50000	     28622 ns/op	    1916 B/op	      10 allocs/op
PASS
```
with one args and is
```
5 args
30000	     46051 ns/op	    3014 B/op	      21 allocs/op
PASS

10 args
20000	     70161 ns/op	    4465 B/op	      35 allocs/op
PASS
```
with multiple args.

The result with `o.outC <- stringify // := []string{}`. It is
```
50000	     36621 ns/op	    2346 B/op	      11 allocs/op
PASS
```
with one args and is
```
5args
20000	     73260 ns/op	    4494 B/op	      28 allocs/op
PASS
```
with multiple args.

The result with `o.outC <- buffer.String()`.  It is
```
50000	     35578 ns/op	    2401 B/op	      13 allocs/op
PASS
```
with one args and is
```
5args
20000	     72620 ns/op	    4455 B/op	      24 allocs/op
PASS
```
with multiple args.