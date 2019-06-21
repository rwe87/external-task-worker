/*
 * Copyright 2019 InfAI (CC SES)
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
	"encoding/json"
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
	"log"
	"net/url"
	"sync"

	"errors"

	"github.com/SENERGY-Platform/external-task-worker/util"

)

type Iot struct {
	cache *Cache
	url   string
}

var iot *Iot
var once sync.Once

func GetIot() *Iot {
	once.Do(func() {
		iot = NewIot(util.Config.DeviceRepoUrl)
	})
	return iot
}

func NewIot(url string) (*Iot){
	return &Iot{url: url, cache:NewCache()}
}

func (this *Iot) GetDeviceInstance(token JwtImpersonate, deviceInstanceId string) (result model.DeviceInstance, err error) {
	if err = this.CheckExecutionAccess(token, deviceInstanceId); err == nil {
		result, err = this.getDeviceFromCache(deviceInstanceId)
		if err != nil {
			err = token.GetJSON(this.url+"/devices/"+url.QueryEscape(deviceInstanceId), &result)
		}
	}
	return
}

func (this *Iot) getDeviceFromCache(id string) (device model.DeviceInstance, err error) {
	item, err := this.cache.Get("device."+id)
	if err != nil {
		return device, err
	}
	err = json.Unmarshal(item.Value, &device)
	return
}

func (this *Iot) getServiceFromCache(id string) (service model.Service, err error) {
	item, err := this.cache.Get("service."+id)
	if err != nil {
		return service, err
	}
	err = json.Unmarshal(item.Value, &service)
	return
}

func (this *Iot) GetDeviceService(token JwtImpersonate, serviceId string) (result model.Service, err error) {
	result, err = this.getServiceFromCache(serviceId)
	if err != nil {
		err = token.GetJSON(this.url+"/services/"+url.QueryEscape(serviceId), &result)
	}
	return
}

func (this *Iot) CheckExecutionAccess(token JwtImpersonate, deviceId string) (err error) {
	result, err := this.getAccessFromCache(token, deviceId)
	if err != nil {
		err = token.GetJSON(util.Config.PermissionsUrl + "/jwt/check/deviceinstance/" + url.QueryEscape(deviceId) + "/x/bool", &result)
	}
	if err != nil {
		return err
	}
	if result {
		return nil
	}else{
		return errors.New("user may not execute events for the resource")
	}
}

func (this *Iot) getAccessFromCache(token JwtImpersonate, id string) (result bool, err error) {
	item, err := this.cache.Get("check.device."+string(token)+"."+id)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(item.Value, &result)
	return
}

func (this *Iot)  GetDeviceInfo(instanceId string, serviceId string, user string) (instance model.DeviceInstance, service model.Service, err error) {
	token, err := GetUserToken(user)
	if err != nil {
		log.Println("error on user token generation: ", err)
		return instance, service, err
	}
	instance, err = this.GetDeviceInstance(token, instanceId)
	if err != nil {
		log.Println("error on getDeviceInfo GetDeviceInstance")
		return
	}
	service, err = this.GetDeviceService(token, serviceId)
	return
}


func GetDeviceInfo(instanceId string, serviceId string, user string) (instance model.DeviceInstance, service model.Service, err error) {
	return GetIot().GetDeviceInfo(instanceId, serviceId, user)
}
