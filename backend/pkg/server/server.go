package server

import (
    "backend/pkg/database"
    "encoding/json"
    "fmt"
    "github.com/brentp/go-chartjs"
    "github.com/brentp/go-chartjs/types"
    "github.com/gorilla/mux"
    "log"
    "math"
    "net/http"
    "strconv"
)

type Server struct {
    db *database.DB
}

func NewServer(db *database.DB) *Server {
    return &Server{db: db}
}

func write(w http.ResponseWriter, status int, payload interface{}) error {
    response, err := json.Marshal(payload)
    if err != nil {
        log.Printf("Faled to marshal payload: %v", err)
        writeError(w, err)
        return err
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if _, err := w.Write([]byte(response)); err != nil {
        log.Printf("Failed to write response: %v", err)
        return err
    }
    return nil
}

func writeError(w http.ResponseWriter, err error) {
    w.WriteHeader(http.StatusInternalServerError)
    if _, err := w.Write([]byte(err.Error())); err != nil {
        log.Printf("Faled write error: %v", err)
    }
}

func (s *Server) workcenters(w http.ResponseWriter, r *http.Request) {
    ws, err := s.db.Workcenters()
    if err != nil {
        writeError(w, err)
        return
    }
    seqs, err := s.db.WorkcenterSequenceCount()
    if err != nil {
        writeError(w, err)
        return
    }
    answer := map[string]interface{}{
        "workcenters": ws,
        "sequences": seqs,
    }
    _ = write(w, http.StatusOK, answer)
}

func (s *Server) weeklyOvertime(w http.ResponseWriter, r *http.Request) {
    ot, err := s.db.WorkcenterOvertimeWeekly()
    if err != nil {
        writeError(w, err)
        return
    }

    // TODO: Graphing/plotting
    answer := map[string]interface{}{
        "overtime": ot,
    }
    _ = write(w, http.StatusOK, answer)
}

func (s *Server) dailyOvertime(w http.ResponseWriter, r *http.Request) {
    ot, err := s.db.WorkcenterOvertimeWeekly()
    if err != nil {
        writeError(w, err)
        return
    }

    // TODO: Graphing/plotting
    answer := map[string]interface{}{
        "overtime": ot,
    }
    _ = write(w, http.StatusOK, answer)
}

func (s *Server) specialOrders(w http.ResponseWriter, r *http.Request) {
    so, err := s.db.SpecialOrders()
    if err != nil {
        writeError(w, err)
        return
    }

    // TODO: Graphing/plotting
    answer := map[string]interface{}{
        "orders": so,
    }
    _ = write(w, http.StatusOK, answer)
}

func (s *Server) durationLowerBound(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    orderID, _ := strconv.ParseInt(vars["order_id"], 10, 64)
    so, err := s.db.DurationLowerBound(orderID)
    if err != nil {
        writeError(w, err)
        return
    }

    // TODO: Graphing/plotting
    answer := map[string]interface{}{
        "orders": so,
    }
    _ = write(w, http.StatusOK, answer)
}

func (s *Server) missedOrderDeadlines(w http.ResponseWriter, r *http.Request) {
    mod, err := s.db.MissedOrderDeadlines()
    if err != nil {
        writeError(w, err)
        return
    }

    // TODO: Graphing/plotting
    answer := map[string]interface{}{
        "orders": mod,
    }
    _ = write(w, http.StatusOK, answer)
}

type xy struct {
    x []float64
    y []float64
    r []float64
}

func (v xy) Xs() []float64 {
    return v.x
}
func (v xy) Ys() []float64 {
    return v.y
}
func (v xy) Rs() []float64 {
    return v.r
}

func check(e error) {
    if e != nil {
        log.Fatal(e)
    }
}

func (s *Server) orderHistory(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    orderID, _ := vars["order_id"]
    oh, err := s.db.OrderHistory(orderID)
    if err != nil {
        writeError(w, err)
        return
    }
    if r.FormValue("plot") == "true" {
        var xys1 xy
        var xys2 xy

        // make some example data.
        for i := float64(0); i < 9; i += 0.1 {
            xys1.x = append(xys1.x, i)
            xys2.x = append(xys2.x, i)

            xys1.y = append(xys1.y, math.Sin(i))
            xys2.y = append(xys2.y, 3*math.Cos(2*i))

        }

        // a set of colors to work with.
        colors := []*types.RGBA{
            &types.RGBA{102, 194, 165, 220},
            &types.RGBA{250, 141, 98, 220},
            &types.RGBA{141, 159, 202, 220},
            &types.RGBA{230, 138, 195, 220},
        }

        // a Dataset contains the data and styling info.
        d1 := chartjs.Dataset{Data: xys1, BorderColor: colors[1], Label: "sin(x)", Fill: chartjs.False,
            PointRadius: 10, PointBorderWidth: 4, BackgroundColor: colors[0]}

        d2 := chartjs.Dataset{Data: xys2, BorderWidth: 8, BorderColor: colors[3], Label: "3*cos(2*x)",
            Fill: chartjs.False, PointStyle: chartjs.Star}

        chart := chartjs.Chart{Label: "test-chart"}

        var err error
        _, err = chart.AddXAxis(chartjs.Axis{Type: chartjs.Linear, Position: chartjs.Bottom, ScaleLabel: &chartjs.ScaleLabel{FontSize: 22, LabelString: "X", Display: chartjs.True}})
        check(err)
        d1.YAxisID, err = chart.AddYAxis(chartjs.Axis{Type: chartjs.Linear, Position: chartjs.Left,
            ScaleLabel: &chartjs.ScaleLabel{LabelString: "sin(x)", Display: chartjs.True}})
        check(err)
        chart.AddDataset(d1)

        d2.YAxisID, err = chart.AddYAxis(chartjs.Axis{Type: chartjs.Linear, Position: chartjs.Right,
            ScaleLabel: &chartjs.ScaleLabel{LabelString: "3*cos(2*x)", Display: chartjs.True}})
        check(err)
        chart.AddDataset(d2)

        chart.Options.Responsive = chartjs.False

        if err := chart.SaveHTML(w, nil); err != nil {
            fmt.Println("PANIC!")
            return
        }

        return
    }

    answer := map[string]interface{}{
        "orders": oh,
    }
    _ = write(w, http.StatusOK, answer)
}
