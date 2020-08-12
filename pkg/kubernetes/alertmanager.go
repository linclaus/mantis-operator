package kubernetes

import (
	"fmt"

	alertmangerconfig "github.com/prometheus/alertmanager/config"
)

func UpdatedReceivers(rvs []*alertmangerconfig.Receiver, strategyId string) []*alertmangerconfig.Receiver {
	var rv *alertmangerconfig.Receiver
	index := -1
	for i, receive := range rvs {
		fmt.Println(receive)
		if receive.Name == strategyId {
			index = i
			break
		}
	}
	rv = &alertmangerconfig.Receiver{
		Name: strategyId,
	}
	if index == -1 {
		return append(rvs, rv)
	} else {
		rvs[index] = rv
		return rvs
	}
}

func DeletedReceivers(rvs []*alertmangerconfig.Receiver, strategyId string) []*alertmangerconfig.Receiver {
	index := -1
	for i, receive := range rvs {
		fmt.Println(receive)
		if receive.Name == strategyId {
			index = i
			break
		}
	}

	if index != -1 {
		return append(rvs[:index], rvs[index+1:]...)
	} else {
		return rvs
	}
}

func UpdatedRoutes(rts []*alertmangerconfig.Route, strategyId string) []*alertmangerconfig.Route {
	var rt *alertmangerconfig.Route
	index := -1
	for i, route := range rts {
		fmt.Println(route)
		if route.Receiver == strategyId {
			index = i
			break
		}
	}
	rt = &alertmangerconfig.Route{
		Receiver: strategyId,
	}
	if index == -1 {
		return append(rts, rt)
	} else {
		rts[index] = rt
		return rts
	}
}

func DeletedRoutes(rts []*alertmangerconfig.Route, strategyId string) []*alertmangerconfig.Route {
	index := -1
	for i, route := range rts {
		fmt.Println(route)
		if route.Receiver == strategyId {
			index = i
			break
		}
	}
	if index != -1 {
		return append(rts[:index], rts[index+1:]...)
	} else {
		return rts
	}
}
