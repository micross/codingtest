package worker

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/micross/codingtest/models"
	"github.com/micross/codingtest/utils"
)

func Work() {

	ch, err := models.RabbitMQ.Channel()
	utils.FailOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"codingtest", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	utils.FailOnError(err, "Failed to declare a queue")

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	utils.FailOnError(err, "Failed to set QoS")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	utils.FailOnError(err, "Failed to register a consumer")

	forever := make(chan bool)
	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			process(string(d.Body))
			d.Ack(false)
		}
	}()

	<-forever
}

func process(data string) {
	msg := strings.Split(data, "#")
	fromAccountID := msg[0]
	toAccountID := msg[1]
	name := msg[2]
	journalID := msg[3]
	flag := msg[4]

	var toAccount models.Account
	if err := models.DB.Where("id = ?", toAccountID).Where("status = ?", 1).First(&toAccount).Error; err != nil {
		failure(fromAccountID, toAccountID, journalID)
		return
	}

	var toOwner models.Owner
	if err := models.DB.Where("id = ?", toAccount.OwnerId).First(&toOwner).Error; err != nil {
		failure(fromAccountID, toAccountID, journalID)
		return
	}

	if toOwner.Name != name {
		failure(fromAccountID, toAccountID, journalID)
		return
	}

	if flag == "2" && doRemoteRequest() != true {
		failure(fromAccountID, toAccountID, journalID)
		return
	}

	var journal models.Journal
	if err := models.DB.Where("id = ?", journalID).Where("status = ?", 2).First(&journal).Error; err != nil {
		failure(fromAccountID, toAccountID, journalID)
		return
	}

	toAccount.Balance = toAccount.Balance + journal.Amount

	tx := models.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return
	}

	if err := tx.Save(&toAccount).Error; err != nil {
		tx.Rollback()
		return
	}

	journal.Status = 1
	if err := tx.Save(&journal).Error; err != nil {
		tx.Rollback()
		return
	}

	if err := tx.Commit().Error; err != nil {
		failure(fromAccountID, toAccountID, journalID)
		return
	}
	return
}

func failure(fromAccountID, toAccountID, journalID string) {
	var fromAccount models.Account
	if err := models.DB.Where("id = ?", fromAccountID).Where("status = ?", 1).First(&fromAccount).Error; err != nil {
		return
	}

	var journal models.Journal
	if err := models.DB.Where("id = ?", journalID).Where("status = ?", 2).First(&journal).Error; err != nil {
		return
	}

	fromAccount.Balance = fromAccount.Balance + journal.Amount + journal.Charge

	tx := models.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return
	}

	if err := tx.Save(&fromAccount).Error; err != nil {
		tx.Rollback()
		return
	}

	journal.Status = 3
	if err := tx.Save(&journal).Error; err != nil {
		tx.Rollback()
		return
	}

	if err := tx.Commit().Error; err != nil {
		return
	}

	RedisConn := models.RedisPool.Get()
	defer RedisConn.Close()

	RedisConn.Do("DECRBY", fromAccountID, journal.Amount)
}

func doRemoteRequest() bool {
	resp, err := http.Get("http://handy.travel/test/success.json")
	if err != nil {
		return false
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	type Result struct {
		Status string
	}
	var r Result
	err = json.Unmarshal(body, &r)
	if err != nil {
		return false
	}

	return r.Status == "success"
}
