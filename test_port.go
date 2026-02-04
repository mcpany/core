package main
import (
    "fmt"
    "net"
)
func main() {
    l, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil { panic(err) }
    fmt.Println(l.Addr().String())
    fmt.Println(l.Addr().(*net.TCPAddr).Port)
}
