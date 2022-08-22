package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type storage struct{
	sync.RWMutex
	collector map[string]chan string
}

func (s *storage) put_handler(w http.ResponseWriter, r *http.Request){

	key := r.URL.Path
	value := r.URL.Query().Get("v")

	if value == ""{
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if_channels_exists(s, key)
	go add_to_channel(s, key, value)
}

func (s *storage) get_handler(w http.ResponseWriter, r *http.Request){

	key := r.URL.Path
	var timeout int
	timeout_string := r.URL.Query().Get("timeout")

	if timeout_string!= ""{
		var err error
		timeout, err = strconv.Atoi(timeout_string)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}

	if_channels_exists(s, key)
	tmp := s.collector[key]

	select {
		case msg := <- tmp:
			fmt.Fprint(w, msg)
		case <- time.After(time.Second * time.Duration(timeout)):
			w.WriteHeader(http.StatusNotFound)
	}
}

func if_channels_exists(s *storage, key string){
	if _, ok := s.collector[key]; !ok {
		s.collector[key] = make (chan string)
	}
}

func add_to_channel(s *storage, key, value string){

	s.RLock()
	tmp := s.collector[key]
	s.RUnlock()
	tmp <- value
}

func Make_storage() *storage{

	var s storage

	s.collector = make(map[string]chan string)
	return &s
}

func Method_separator(w http.ResponseWriter, r *http.Request, s *storage){

	if r.Method == "PUT"{
		s.put_handler(w,r)
	} else if r.Method == "GET"{
		s.get_handler(w,r)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func handlefunc(port string){

	s := Make_storage()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    		Method_separator(w, r, s)
		})
	http.ListenAndServe("127.0.0.1:"+port, nil)

}	

func main(){
	
	var port string

	flag.StringVar(&port, "port", "8080", "choose port")
	flag.Parse()
	if _, err := strconv.Atoi(port); err != nil && len(port) != 4{
		log.Fatal("port is not a number or len bigget than 4", port)
	}
	handlefunc(port)
}