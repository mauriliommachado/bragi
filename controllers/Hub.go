package controllers

import (
	"github.com/gomodels"
	"gopkg.in/mgo.v2/bson"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/go-redis/redis"
	"fmt"
	"log"
)

type Hub struct {
}

func Run(){
	client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
}

var client *redis.Client

type UserConn struct {
	user models.User
	conn *websocket.Conn
}

type Channel struct {
	id    bson.ObjectId
	users *[]UserConn
}

var channels []Channel

func NewHub() (Hub) {
	return Hub{}
}

func (hub *Hub) Register(user models.User, conn *websocket.Conn) {
	userConn := UserConn{user: user, conn: conn}
	for _, id := range user.Rooms {
		channel, err := findChannelById(id)
		if err != nil {
			*channel = Channel{id: id, users: &[]UserConn{userConn}}
			channels = append(channels, *channel)
			log.Println("adicionado canal", channel.id.Hex())
			openListener(*channel)
		} else {
			*channel.users = append(*channel.users, userConn)
		}
	}
}

func findChannelById(id bson.ObjectId) (*Channel, error) {
	for _, val := range channels {
		if val.id == id {
			return &val, nil
		}
	}
	return &Channel{}, errors.New("não encontrado")
}

func (hub *Hub) UnRegister(user models.User) {
	for _, val := range user.Rooms {
		channel, _ := findChannelById(val)
		log.Println("retirando usuário",user.Name,"do canal", channel.id.Hex())
		for i, valUser := range *channel.users {
			if valUser.user.Id == user.Id {
				valUser.conn.Close()
				valUser.conn = nil
				users := append((*channel.users)[:i], (*(*channel).users)[i+1:]...)
				*channel.users = users
				break
			}
		}
	}
}

func openListener(channel Channel){
	log.Println("abrindo escutador para o canal",channel.id.Hex())
	go Listener(client.Subscribe(channel.id.Hex()))
}

func Listener(pubChannel *redis.PubSub) {
	for {
		var msgi, err interface{} = pubChannel.ReceiveMessage()
		if err != nil {
			break
		}
		switch msg := msgi.(type) {
		case *redis.Subscription:
			fmt.Println("subscribed to", msg.Channel)
		case *redis.Message:
			fmt.Println("received", msg.Payload, "from", msg.Channel)
			channel, _ := findChannelById(bson.ObjectIdHex(msg.Channel))
			for _, valUser := range *channel.users {
				fmt.Println("sendind to", valUser.user.Name)
				valUser.conn.WriteMessage(websocket.TextMessage,[]byte(msg.Payload))
			}
		default:
			panic(fmt.Errorf("unknown message: %#v", msgi))
		}
	}
}
