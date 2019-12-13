package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.ReadInConfig()
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetDefault("port", "8080")
	viper.SetDefault("origin", "*")

	mongoHost := viper.GetString("mongo.host")
	mongoUser := viper.GetString("mongo.user")
	mongoPass := viper.GetString("mongo.pass")
	mongoDB := viper.GetString("mongo.db")
	mongoCollection := viper.GetString("mongo.collection")
	origin := viper.GetString("origin")
	port := fmt.Sprintf(":%v", viper.GetString("port"))

	connString := fmt.Sprintf("%v:%v@%v", mongoUser, mongoPass, mongoHost)
	conn, err := mgo.Dial(connString)
	if err != nil {
		log.Printf("dial mongodb server with connection string %q: %v", connString, err)
		return
	}

	h := &handler{
		mongo: conn,
		db:    mongoDB,
		col:   mongoCollection,
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{origin},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodOptions},
	}))

	e.GET("/records", h.list)
	e.POST("/records", h.create)
	e.DELETE("/records/:id", h.remove)

	e.Logger.Fatal(e.Start(port))
}

type Member struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Record struct {
	ObjectID bson.ObjectId `json:"_id" bson:"_id"`
	ID       string        `json:"id"`
	Amount   int           `json:"amount"`
	Payer    Member        `json:"payer"`
	Owes     []Member      `json:"owes"`
	Note     string        `json:"note"`
}

type handler struct {
	mongo *mgo.Session
	db    string
	col   string
}

func (h *handler) list(c echo.Context) error {
	conn := h.mongo.Copy()
	defer conn.Close()
	ts := []Record{}
	if err := conn.DB(h.db).C(h.col).Find(nil).All(&ts); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, ts)
}

func (h *handler) create(c echo.Context) error {
	id := bson.NewObjectId()
	var t Record
	if err := c.Bind(&t); err != nil {
		return err
	}
	t.ObjectID = id

	// TODO: synchronized
	// t.Done = false

	conn := h.mongo.Copy()
	defer conn.Close()
	if err := conn.DB(h.db).C(h.col).Insert(t); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, t)
}

func (h *handler) remove(c echo.Context) error {
	conn := h.mongo.Copy()
	defer conn.Close()
	id := bson.ObjectIdHex(c.Param("id"))

	if err := conn.DB(h.db).C(h.col).RemoveId(id); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"_id": id,
	})
}
