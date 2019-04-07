package server

import (
    "github.com/gorilla/mux"
    "net/http"
)

func NewRouter(server *Server, assets string) *mux.Router {
    router := mux.NewRouter().StrictSlash(true)

    // Server CSS, JS & Images Statically.
    router.Handle("/", http.FileServer(http.Dir("."+assets)))

    router.HandleFunc("/workcenters", server.workcenters)
    router.HandleFunc("/workcenters/overtime/weekly", server.weeklyOvertime)
    router.HandleFunc("/workcenters/overtime/daily", server.dailyOvertime)
    router.HandleFunc("/orders/missed", server.missedOrderDeadlines)
    router.HandleFunc("/orders/special", server.specialOrders)
    router.HandleFunc("/orders/{order_id:[0-9]+}/history", server.orderHistory)
    router.HandleFunc("/orders/{order_id:[0-9]+}/duration", server.durationLowerBound)
    router.PathPrefix(assets).Handler(http.StripPrefix(assets, http.FileServer(http.Dir("."+assets))))

    return router
}