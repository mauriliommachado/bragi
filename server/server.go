package server

import (
	"net/http"
	"log"
	"github.com/gorilla/websocket"
	"flag"
	"github.com/gomodels"
	"../controllers"
)

type ServerProperties struct {
	Port    string
	Address string
}

var hub = controllers.NewHub()

var mc = controllers.NewMessageController()

func validAuthHeader(req *http.Request) (bool,models.User) {
	auth := req.URL.Query().Get("token")
	var user models.User
	if len(auth) < 0 {
		return false,user
	}
	user.Token = auth
	if mc.Validate(&user){
		return true, user
	}else{
		return false, user
	}
}

var addr = flag.String("addr", "localhost:8083", "http service address")

var upgrader = websocket.Upgrader{CheckOrigin:func(r *http.Request) bool {
	return true
}}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
}

func echo(w http.ResponseWriter, r *http.Request) {
	validou, user := validAuthHeader(r)
	if !validou {
		unauthorized(w)
		return
	}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	log.Println(user.Name, "has logged in")
	go hub.Register(user, c)
	go func(user models.User) {
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				go hub.UnRegister(user)
				break
			}

			log.Printf("recv: %s", message)
			err = c.WriteMessage(mt, message)
			if err != nil {
				log.Println("write:", err)
				hub.UnRegister(user)
				break
			}
		}
	}(user)
}


func Start(properties ServerProperties) {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/echo", echo)
	log.Fatal(http.ListenAndServe(*addr, nil))

}

