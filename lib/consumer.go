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
	"github.com/SENERGY-Platform/iot-broker-client-lib"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"log"
)

var consumer *iot_broker_client_lib.Consumer

func InitConsumer() (err error) {
	consumer, err = iot_broker_client_lib.NewConsumer(util.Config.AmqpUrl, util.Config.ConsumerName, util.Config.ResponseTopic, false, int(util.Config.AmqpPrefetchCount), func(msg []byte) error {
		return CompleteCamundaTask(string(msg))
	})
	if err != nil {
		log.Println("ERROR: unable to init amqp connection", err)
		return err
	}
	err = consumer.BindAll()
	if err != nil {
		log.Println("ERROR: unable to bind devices to consumer", err)
		return err
	}
	return
}

func CloseConsumer(){
	consumer.Close()
}