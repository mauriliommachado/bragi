package controllers

import (
	"net/http"
	"github.com/gomodels"
	"log"
	"gopkg.in/mgo.v2/bson"
	"encoding/json"
	"errors"
)

type MessageController struct {
}

func NewMessageController() MessageController {
	return MessageController{}
}

func (rc *MessageController) FindUser(token string, id bson.ObjectId) (models.User, error) {
	var req *http.Request
	var user models.User

	req, err := http.NewRequest(http.MethodGet, models.ID_MS_URL+"/"+id.Hex(), nil)
	req.Header.Set("Authorization", "Basic "+token)
	if err != nil {
		log.Println(err)
		return user, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return user, err
	}

	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		log.Println(err)
		return user, err
	}
	if len(user.Id.Hex()) == 0 {
		err = errors.New("Usuário não encontrado")
		log.Println(err)
		return user, err
	}
	return user, nil
}

func (rc *MessageController) Validate(user *models.User) bool {
	req, err := http.NewRequest(http.MethodGet, models.ID_MS_URL+"/validate/"+user.Token, nil)
	if err != nil {
		log.Println(err)
		return false
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		log.Println(err)
		return false
	}
	if len(user.Id.Hex()) == 0 {
		err = errors.New("Usuário não encontrado")
		log.Println(err)
		return false
	}
	return true
}
