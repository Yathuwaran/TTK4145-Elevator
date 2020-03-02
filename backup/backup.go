package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "strconv"
)



type Elevator struct {
    Elevator[]Status `json:"elevator"`
}

type Status struct {
    ID  int `json:"id"`
    Motor_dir   int `json:"dir"`
    Queue  [][]int    `json:"queue"`
    Current_floor int `json:"current_floor"`

}


func read_backup() Elevator {
    // Open our jsonFile
    jsonFile, err := os.Open("elevator.json")
 
    if err != nil {
        fmt.Println(err)
    }

    fmt.Println("Successfully Opened Elevator.json")
    defer jsonFile.Close()

    // read our opened xmlFile as a byte array.
    byteValue, _ := ioutil.ReadAll(jsonFile)

    var elevator Elevator

    json.Unmarshal(byteValue, &elevator)

    //Accessing json-struct elevator.Elevator[0]

    for i := 0; i < len(elevator.Elevator); i++ {


        fmt.Println("Elevator ID: " + strconv.Itoa(elevator.Elevator[i].ID))
        fmt.Println("Motor direction: " + strconv.Itoa(elevator.Elevator[i].Motor_dir))
        fmt.Println("Queue:")
        fmt.Println(elevator.Elevator[i].Queue)
        fmt.Println("Current Floor: " + strconv.Itoa(elevator.Elevator[i].Current_floor))
    }

    return elevator

}


func write_backup(elevator Elevator) {
  
    data, _ := json.MarshalIndent(elevator, "", " ")

    //fmt.Println(string(b))
    // writing json to file

    _ = ioutil.WriteFile("elevator.json", data, 0644)

      
}

func main(){
    var elev Elevator

    elev = read_backup()
    write_backup(elev) 
}