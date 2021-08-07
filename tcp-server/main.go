package main

import (
	"bufio"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/protobuf/proto"
)

const (
	connHost = "localhost"
	connPort = "8080"
	connType = "tcp"
)

var (
	getUser        *sql.Stmt
	updateNickname *sql.Stmt
)

func main() {

	db, err := sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/entry_task")
	if err != nil {
		log.Fatal("Unable to connect with mysql: ", err)
	}
	defer db.Close()

	getUser, err = db.Prepare("SELECT nickname, password FROM user WHERE username=?")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer getUser.Close()

	updateNickname, err = db.Prepare("UPDATE user SET nickname = ? WHERE username = ?")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer updateNickname.Close()

	log.Print(GetMD5Hash("ABC"))

	// Open socket connection
	fmt.Println("Starting " + connType + " server on " + connHost + ":" + connPort)
	l, err := net.Listen(connType, connHost+":"+connPort)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println("Error connecting:", err.Error())
			return
		}
		fmt.Println("Client " + c.RemoteAddr().String() + " connected.")
		go handleConnection(c)
	}
}

func handleConnection(conn net.Conn) {
	buffer, err := bufio.NewReader(conn).ReadBytes(0) // stop when /x00
	if err != nil {
		fmt.Println("Client left.")
		conn.Close()
		return
	}
	res, err := handleRequest(buffer[:len(buffer)-1])
	if err != nil {
		fmt.Println("Request failed.")
		conn.Write([]byte{0})
	} else {
		conn.Write(append(res, []byte{0}...))
	}

	handleConnection(conn)
}

func handleRequest(in []byte) ([]byte, error) {
	req := &Request{}
	res := &Request{}
	if err := proto.Unmarshal(in, req); err != nil {
		log.Print("Failed to parse request from client: ", err)
		res.ResponseType = Request_FAILED
	} else {
		switch req.RequestType {
		case Request_AUTH:
			var nickname, password string
			err = getUser.QueryRow(req.Username).Scan(&nickname, &password)
			if err != nil {
				log.Print("AUTH failed: ", err)
				res.ResponseType = Request_FAILED
			} else if GetMD5Hash(req.Password) == password {
				res.ResponseType = Request_SUCCESS
				res.Nickname = nickname
				res.Username = req.Username
				res.Password = ""
			} else {
				log.Print("AUTH failed: Password incorrect")
				res.ResponseType = Request_FAILED
			}
		case Request_EDIT_NICKNAME:
			_, err = updateNickname.Exec(req.Nickname, req.Username)
			if err != nil {
				log.Print("EDIT_NICKNAME failed: ", err)
				res.ResponseType = Request_FAILED
			} else {
				res.ResponseType = Request_SUCCESS
				res.Username = req.Username
				res.Nickname = req.Nickname
				res.Password = ""
			}
		default:
			res.ResponseType = Request_FAILED
		}
	}
	out, err := proto.Marshal(res)
	if err != nil {
		log.Print("Failed to parse response: ", err)
		return nil, err
	}
	log.Print("Handle request: ", res)
	return out, nil
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
