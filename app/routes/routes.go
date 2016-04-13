// GENERATED CODE - DO NOT EDIT
package routes

import "github.com/revel/revel"


type tHome struct {}
var Home tHome


func (_ tHome) Index(
		) string {
	args := make(map[string]string)
	
	return revel.MainRouter.Reverse("Home.Index", args).Url
}


type tKiteQ struct {}
var KiteQ tKiteQ


func (_ tKiteQ) Kiteqs(
		apName string,
		) string {
	args := make(map[string]string)
	
	revel.Unbind(args, "apName", apName)
	return revel.MainRouter.Reverse("KiteQ.Kiteqs", args).Url
}

func (_ tKiteQ) DelSubscribe(
		apName string,
		group string,
		topic string,
		) string {
	args := make(map[string]string)
	
	revel.Unbind(args, "apName", apName)
	revel.Unbind(args, "group", group)
	revel.Unbind(args, "topic", topic)
	return revel.MainRouter.Reverse("KiteQ.DelSubscribe", args).Url
}

func (_ tKiteQ) MinuteChart(
		apName string,
		end string,
		) string {
	args := make(map[string]string)
	
	revel.Unbind(args, "apName", apName)
	revel.Unbind(args, "end", end)
	return revel.MainRouter.Reverse("KiteQ.MinuteChart", args).Url
}

func (_ tKiteQ) HourChart(
		apName string,
		end string,
		) string {
	args := make(map[string]string)
	
	revel.Unbind(args, "apName", apName)
	revel.Unbind(args, "end", end)
	return revel.MainRouter.Reverse("KiteQ.HourChart", args).Url
}


type tJobs struct {}
var Jobs tJobs


func (_ tJobs) Status(
		) string {
	args := make(map[string]string)
	
	return revel.MainRouter.Reverse("Jobs.Status", args).Url
}


type tStatic struct {}
var Static tStatic


func (_ tStatic) Serve(
		prefix string,
		filepath string,
		) string {
	args := make(map[string]string)
	
	revel.Unbind(args, "prefix", prefix)
	revel.Unbind(args, "filepath", filepath)
	return revel.MainRouter.Reverse("Static.Serve", args).Url
}

func (_ tStatic) ServeModule(
		moduleName string,
		prefix string,
		filepath string,
		) string {
	args := make(map[string]string)
	
	revel.Unbind(args, "moduleName", moduleName)
	revel.Unbind(args, "prefix", prefix)
	revel.Unbind(args, "filepath", filepath)
	return revel.MainRouter.Reverse("Static.ServeModule", args).Url
}


type tTestRunner struct {}
var TestRunner tTestRunner


func (_ tTestRunner) Index(
		) string {
	args := make(map[string]string)
	
	return revel.MainRouter.Reverse("TestRunner.Index", args).Url
}

func (_ tTestRunner) Run(
		suite string,
		test string,
		) string {
	args := make(map[string]string)
	
	revel.Unbind(args, "suite", suite)
	revel.Unbind(args, "test", test)
	return revel.MainRouter.Reverse("TestRunner.Run", args).Url
}

func (_ tTestRunner) List(
		) string {
	args := make(map[string]string)
	
	return revel.MainRouter.Reverse("TestRunner.List", args).Url
}


