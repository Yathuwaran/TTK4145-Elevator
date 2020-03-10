package backup

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    ."../structs"
)


func Read_backup() Message_struct {
    // Open our jsonFile
    jsonFile, err := os.Open("elevator.json")
    if err != nil {
        fmt.Println(err)
    }

    fmt.Println("Successfully Opened Elevator.json")
    defer jsonFile.Close()

    // read our opened json as a byte array.
    byteValue, _ := ioutil.ReadAll(jsonFile)
    var elevator Message_struct
    json.Unmarshal(byteValue, &elevator)

    return elevator
}


func Write_backup(elevator Message_struct) {

    data, _ := json.MarshalIndent(elevator, "", " ")
    //fmt.Println(string(b))
    // writing json to file
    _ = ioutil.WriteFile("elevator.json", data, 0644)

}

/*
func main(){

  var mld Message_struct
  mld.ID = "Elevator_1"
  mld.Destination = Order{1,1}
  mld.Last_floor = 2
  mld.Dir = 1
  mld. State = 2
  mld.Queue = [][]int{{0, 1},{6, 7}}
    var elev Message_struct
    elev = backup.Read_backup()
    fmt.Println(elev)
    backup.Write_backup(mld)
}*/
