# Javascript interpreter

``` sh
root@spos# js
>>> console.log("hello world")
hello world
>>> reg = new RegExp("hello")
>>> console.log(reg)
/hello/
>>> resp = reg.test("hello world")
>>> console.log(resp)
true
>>> var url = "http://baidu.com"
>>> resp = http.Get(url)
>>> console.log(resp)
<html>
<meta http-equiv="refresh" content="0;url=http://www.baidu.com/">
</html>
```

# Mount samba filesystem

``` sh
root@spos# mount smb://icexin:spos@172.28.90.3:445/sambashare /share
root@spos# cd /share
root@spos# ls
-rw-rw-rw- 111 fib.js
root@spos# cat fib.js
function fib(n) {
        if (n == 1 || n == 2) {
                return 1;
        }
        return fib(n-1) + fib(n-2);
}

console.log(fib(10))

root@spos# js fib.js
55

```

# Run nes emulator

First run `QEMU_GRAPHIC=true QEMU_ACCEL=true mage graphic`

``` sh
root@spos# ls /share
-rw-rw-rw- 111   fib.js
-rw-rw-rw- 40976 mario.nes
root@spos# nes -rom /share/mario.nes
```

- `W`, `S`, `A`, `D` mapping `up`, `down`, `left` and `right`.
- `K`, `J` mapping `A` and `B`
- `space` and `enter` mapping `select` and `start`
- `Q` to quit game.

# GUI

First run `QEMU_GRAPHIC=true QEMU_ACCEL=true mage graphic`

``` sh
root@spos# uidemo
```


# Chipmunk2D physics engine

First run `QEMU_GRAPHIC=true QEMU_ACCEL=true mage graphic`

``` sh
root@spos# phy
```

# HTTP server

Running a HTTP server in background.

``` sh
root@spos# go httpd
```

visit http://127.0.0.1:8080/debug/pprof in browser