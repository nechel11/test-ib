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
		fmt.Fprint(w, http.StatusBadRequest)
		return
	}

	s.Lock()
	if _, ok := s.collector[key]; !ok {
		s.collector[key] = make (chan string)
	}
	s.Unlock()
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
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, http.StatusBadRequest)
			return
		}
	}

	s.RLock()
	tmp := s.collector[key]
	s.RUnlock()

	select {
		case msg := <- tmp:
			fmt.Fprint(w, msg)
		case <- time.After(time.Second * time.Duration(timeout)):
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, http.StatusBadRequest)
	}
}
func add_to_channel(s *storage, key, value string){
	s.RLock()
	tmp := s.collector[key]
	s.RUnlock()
	tmp <- value
}

func make_storage() *storage{
	var s storage

	s.collector = make(map[string]chan string)
	return &s
}

func method_separator(w http.ResponseWriter, r *http.Request, s *storage){

	if r.Method == "PUT"{
		s.put_handler(w,r)
	} else if r.Method == "GET"{
		s.get_handler(w,r)
	}
}

func handlefunc(port string){
	s := make_storage()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    		method_separator(w, r, s)
		})
	http.ListenAndServe("127.0.0.1:"+port, nil)

}	

func main(){
	var port string

	flag.StringVar(&port, "port", "8080", "choose port")
	flag.Parse()
	if _, err := strconv.Atoi(port); err != nil{
		log.Fatal("port is not a number", port)
	}
	handlefunc(port)
}