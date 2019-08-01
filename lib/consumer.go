/*
 * Copyright 2018 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package lib

import (
	"log"

	"github.com/SENERGY-Platform/external-task-worker/util"

	"time"

	"github.com/wvanbergen/kafka/consumergroup"
	kazoo "github.com/wvanbergen/kazoo-go"
)

func InitConsumer() {
	defer CloseProducer()
	Produce(util.Config.ResponseTopic, "topic_init")

	zk, chroot := kazoo.ParseConnectionString(util.Config.ZookeeperUrl)
	kafkaconf := consumergroup.NewConfig()
	kafkaconf.Consumer.Return.Errors = util.Config.FatalKafkaErrors == "true"
	kafkaconf.Zookeeper.Chroot = chroot
	consumerGroupName := util.Config.KafkaConsumerGroup
	consumer, err := consumergroup.JoinConsumerGroup(
		consumerGroupName,
		[]string{util.Config.ResponseTopic},
		zk,
		kafkaconf)

	if err != nil {
		log.Fatal("error in consumergroup.JoinConsumerGroup()", err)
	}

	defer consumer.Close()

	kafkaTimeout := util.Config.KafkaTimeout
	useTimeout := true
	if kafkaTimeout <= 0 {
		useTimeout = false
		kafkaTimeout = 3600
	}
	kafkaping := time.NewTicker(time.Second * time.Duration(kafkaTimeout/2))
	kafkatimout := time.NewTicker(time.Second * time.Duration(kafkaTimeout))

	timeout := false

	for {
		select {
		case <-kafkaping.C:
			if useTimeout && timeout {
				Produce(util.Config.ResponseTopic, "topic_init")
			}
		case <-kafkatimout.C:
			if useTimeout && timeout {
				log.Fatal("ERROR: kafka missing ping timeout")
			}
			timeout = true
		case errMsg := <-consumer.Errors():
			log.Fatal("kafka consumer error: ", errMsg)
		case msg, ok := <-consumer.Messages():
			if !ok {
				log.Fatal("empty kafka consumer")
			} else {
				if string(msg.Value) != "topic_init" {
					err = CompleteCamundaTask(string(msg.Value))
					if err != nil {
						log.Println("error while processing kafka message", err, string(msg.Value))
					}
				}
				timeout = false
				consumer.CommitUpto(msg)
			}
		}
	}
}
