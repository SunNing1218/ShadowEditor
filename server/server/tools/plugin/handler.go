// Copyright 2017-2020 The ShadowEditor Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
//
// For more information, please visit: https://github.com/tengge1/ShadowEditor
// You can also visit: https://gitee.com/tengge1/ShadowEditor

package plugin

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tengge1/shadoweditor/helper"
	"github.com/tengge1/shadoweditor/server"
)

func init() {
	handler := Plugin{}
	server.Mux.UsingContext().Handle(http.MethodGet, "/api/Plugin/List", handler.List)
	server.Mux.UsingContext().Handle(http.MethodPost, "/api/Plugin/Add", handler.Add)
	server.Mux.UsingContext().Handle(http.MethodPost, "/api/Plugin/Edit", handler.Edit)
	server.Mux.UsingContext().Handle(http.MethodPost, "/api/Plugin/Delete", handler.Delete)
}

// Plugin 插件控制器
type Plugin struct {
}

// List 获取列表
func (Plugin) List(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	pageSize, err := strconv.Atoi(r.FormValue("pageSize"))
	if err != nil {
		pageSize = 20
	}
	pageNum, err := strconv.Atoi(r.FormValue("pageNum"))
	if err != nil {
		pageNum = 1
	}
	keyword := strings.TrimSpace(r.FormValue("keyword"))

	db, err := server.Mongo()
	if err != nil {
		helper.WriteJSON(w, server.Result{
			Code: 300,
			Msg:  err.Error(),
		})
		return
	}

	filter := bson.M{
		"Status": bson.M{
			"$ne": -1,
		},
	}

	if keyword != "" {
		filter1 := bson.M{
			"Name": bson.M{
				"$regex": keyword,
			},
		}
		filter = bson.M{
			"$and": bson.A{
				filter,
				filter1,
			},
		}
	}

	skip := int64(pageSize * (pageNum - 1))
	limit := int64(pageNum)
	opts := options.FindOptions{
		Skip:  &skip,
		Limit: &limit,
		Sort: bson.M{
			"ID": -1,
		},
	}

	total, _ := db.Count(server.PluginCollectionName, filter)
	var docs bson.A
	db.FindAll(server.PluginCollectionName, &docs, &opts)

	rows := []Model{}

	for _, i := range docs {
		doc := i.(primitive.D).Map()
		info := Model{
			ID:          doc["ID"].(primitive.ObjectID).Hex(),
			Name:        doc["Name"].(string),
			Source:      doc["Source"].(string),
			CreateTime:  doc["CreateTime"].(primitive.DateTime).Time(),
			UpdateTime:  doc["UpdateTime"].(primitive.DateTime).Time(),
			Description: doc["Description"].(string),
			Status:      int(doc["Status"].(int32)),
		}
		rows = append(rows, info)
	}

	helper.WriteJSON(w, server.Result{
		Code: 200,
		Msg:  "Get Successfully!",
		Data: map[string]interface{}{
			"total": total,
			"rows":  rows,
		},
	})
}

// Add 添加
func (Plugin) Add(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name := strings.TrimSpace(r.FormValue("Name"))
	if name == "" {
		helper.WriteJSON(w, server.Result{
			Code: 300,
			Msg:  "Name is not allowed to be empty.",
		})
		return
	}
	source := r.FormValue("Source")
	description := r.FormValue("Description")

	db, err := server.Mongo()
	if err != nil {
		helper.WriteJSON(w, server.Result{
			Code: 300,
			Msg:  err.Error(),
		})
		return
	}

	filter := bson.M{
		"Name": name,
	}
	count, _ := db.Count(server.PluginCollectionName, filter)
	if count > 0 {
		helper.WriteJSON(w, server.Result{
			Code: 300,
			Msg:  "The name is already existed.",
		})
		return
	}

	now := time.Now()

	doc := bson.M{
		"ID":          primitive.NewObjectID(),
		"Name":        name,
		"Source":      source,
		"CreateTime":  now,
		"UpdateTime":  now,
		"Description": description,
		"Status":      0,
	}

	db.InsertOne(server.PluginCollectionName, doc)

	helper.WriteJSON(w, server.Result{
		Code: 200,
		Msg:  "Saved successfully!",
	})
}

// Edit 编辑
func (Plugin) Edit(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id, err := primitive.ObjectIDFromHex(strings.TrimSpace(r.FormValue("ID")))
	name := strings.TrimSpace(r.FormValue("Name"))
	source := r.FormValue("Source")
	description := r.FormValue("Description")

	if err != nil {
		helper.WriteJSON(w, server.Result{
			Code: 300,
			Msg:  "ID is not allowed.",
		})
		return
	}

	if name == "" {
		helper.WriteJSON(w, server.Result{
			Code: 300,
			Msg:  "Name is not allowed to be empty.",
		})
		return
	}

	db, err := server.Mongo()
	if err != nil {
		helper.WriteJSON(w, server.Result{
			Code: 300,
			Msg:  err.Error(),
		})
		return
	}

	filter := bson.M{
		"ID": id,
	}
	var doc interface{}
	find, _ := db.FindOne(server.PluginCollectionName, filter, &doc)
	if !find {
		helper.WriteJSON(w, server.Result{
			Code: 300,
			Msg:  "The plugin is not existed.",
		})
		return
	}

	update := bson.M{
		"$set": bson.M{
			"Name":        name,
			"Source":      source,
			"UpdateTime":  time.Now(),
			"Description": description,
		},
	}

	db.UpdateOne(server.PluginCollectionName, filter, update)

	helper.WriteJSON(w, server.Result{
		Code: 200,
		Msg:  "Saved successfully!",
	})
}

// Delete 删除
func (Plugin) Delete(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id, err := primitive.ObjectIDFromHex(strings.TrimSpace(r.FormValue("ID")))
	if err != nil {
		helper.WriteJSON(w, server.Result{
			Code: 300,
			Msg:  "ID is not allowed.",
		})
		return
	}

	db, err := server.Mongo()
	if err != nil {
		helper.WriteJSON(w, server.Result{
			Code: 300,
			Msg:  err.Error(),
		})
		return
	}

	filter := bson.M{
		"ID": id,
	}

	doc := bson.M{}
	find, _ := db.FindOne(server.MapCollectionName, filter, &doc)

	if !find {
		helper.WriteJSON(w, server.Result{
			Code: 300,
			Msg:  "The plugin is not existed!",
		})
		return
	}

	update := bson.M{
		"$set": bson.M{
			"Status": -1,
		},
	}

	db.UpdateOne(server.PluginCollectionName, filter, update)

	helper.WriteJSON(w, server.Result{
		Code: 200,
		Msg:  "Delete successfully!",
	})
}