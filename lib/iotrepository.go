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
	"net/url"

	"errors"

	"github.com/SmartEnergyPlatform/external-task-worker/util"
	"github.com/SmartEnergyPlatform/iot-device-repository/lib/model"
)

func GetDeviceInstance(token JwtImpersonate, deviceInstanceId string) (result model.DeviceInstance, err error) {
	if err = CheckExecutionAccess(token, deviceInstanceId); err == nil {
		err = token.GetJSON(util.Config.IotRepoUrl+"/deviceInstance/"+url.QueryEscape(deviceInstanceId), &result)
	}
	return
}

func GetDeviceService(token JwtImpersonate, serviceId string) (result model.Service, err error) {
	err = token.GetJSON(util.Config.IotRepoUrl+"/service/"+url.QueryEscape(serviceId), &result)
	return
}

func CheckExecutionAccess(token JwtImpersonate, resource string) (err error) {
	resp, err := token.Get(util.Config.PermissionsUrl + "/jwt/check/deviceinstance/" + url.QueryEscape(resource) + "/x")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("user may not execute events for the resource")
	}
	return nil
}
