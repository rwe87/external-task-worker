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

	"github.com/SENERGY-Platform/external-task-worker/lib/messages"

	"github.com/satori/go.uuid"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"github.com/SmartEnergyPlatform/util/http/request"
)

var workerId = uuid.NewV4().String()

func GetCamundaTask() (tasks []messages.CamundaTask, err error) {
	fetchRequest := messages.CamundaFetchRequest{
		WorkerId: workerId,
		MaxTasks: util.Config.CamundaWorkerTasks,
		Topics:   []messages.CamundaTopic{{LockDuration: util.Config.CamundaFetchLockDuration, Name: util.Config.CamundaTopic}},
	}
	err, _, _ = request.Post(util.Config.CamundaUrl+"/external-task/fetchAndLock", fetchRequest, &tasks)
	return
}

func SetCamundaRetry(taskid string) {
	retry := messages.CamundaRetrySetRequest{Retries: 1}
	request.Put(util.Config.CamundaUrl+"/external-task/"+taskid+"/retries", retry, nil)
}

func CamundaError(task messages.CamundaTask, msg string) {
	errorMsg := messages.CamundaError{WorkerId: GetWorkerId(), ErrorMessage: msg, Retries: 0, ErrorDetails: msg}
	log.Println("Send Error to Camunda: ", msg)
	log.Println(request.Post(util.Config.CamundaUrl+"/external-task/"+task.Id+"/failure", errorMsg, nil))
	//this.completeCamundaTask(taskid, this.GetWorkerId(), "error", messages.BpmnMsg{ErrorMsg:msg})
}

func completeCamundaTask(taskId string, workerId string, outputName string, output messages.BpmnMsg) (err error) {
	if workerId == "" {
		workerId = GetWorkerId()
	}

	variables := map[string]messages.CamundaOutput{
		outputName: {
			Value: output,
		},
	}
	completeRequest := messages.CamundaCompleteRequest{WorkerId: workerId, Variables: variables}
	pl := ""
	var code int
	err, pl, code = request.Post(util.Config.CamundaUrl+"/external-task/"+taskId+"/complete", completeRequest, nil)
	if code == 204 || code == 200 {
		log.Println("complete camunda task: ", completeRequest, pl)
	}else{
		CamundaError(messages.CamundaTask{Id:taskId}, pl)
	}
	return
}

func GetWorkerId() string {
	return workerId
}
