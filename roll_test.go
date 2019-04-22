package main

/*
find files in the pattern of name.ext and roll them back to name.num.ext. Any name.num.ext should also be rolled back by one
*/

import (
    "io/ioutil"
    "testing"
    "os/exec"
    "fmt"
    "os"
)

/******************************************************************************/
// #mark - helpers

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func Tree() {
    cmd := exec.Command("tree", "/tmp/roll_test")
    out, _ := cmd.Output()
    fmt.Printf("Tree:\n%s\n", out)
}

func Write(index int) {
    d1 := []byte( fmt.Sprintf("hello from go %d\n", index) )
    if index==-1 {
        file_name := fmt.Sprintf("/tmp/roll_test/test.txt")
        err := ioutil.WriteFile(file_name, d1, 0644)
        check(err)
    } else {
        file_name := fmt.Sprintf("/tmp/roll_test/test.%d.txt", index)
        err := ioutil.WriteFile(file_name, d1, 0644)
        check(err)
    }
}
/******************************************************************************/
// #mark - setup, tear downs

func init() {
    //fmt.Println("init")
    app_data.file_name = "/tmp/roll_test/test.txt"
}

func setup(older_file_count int) {
    os.Mkdir("/tmp/roll_test", 0770)

    Write(-1)
    
    for i:=0; i<older_file_count; i++ {
        Write(i)
    }
}

func teardown() {
    //fmt.Printf("tear down\n")
    os.RemoveAll("/tmp/roll_test")
}

/******************************************************************************/
// #mark - tests

func TestVprintf(t *testing.T) {
    app_data.verbose = false
    vprintf("")
}

/** this use case is not working yet */
func _TestInitial(t *testing.T) {
    setup(-1)
    
    work()
    
    result := exists("/tmp/roll_test/test.1.txt")
    if result != true {
        t.Error("Expected to find test.1.txt, got ", result)
        Tree()
    }
    
    result = exists("/tmp/roll_test/test.2.txt")
    if result == true {
        t.Error("Expected not to find test.2.txt, got ", result)
        Tree()
    }
    
    teardown()
}

func TestMultiple(t *testing.T) {
    setup(1)
    
    work()
    
    result := true    

    result = exists("/tmp/roll_test/test.0.txt")
    if result != true {
        t.Error("Expected to find test.1.txt, got ", result)
        Tree()
    }
    result = exists("/tmp/roll_test/test.1.txt")
    if result != true {
        t.Error("Expected to find test.1.txt, got ", result)
        Tree()
    }
    
    result = exists("/tmp/roll_test/test.2.test")
    if result == true {
        Tree()
        t.Error("Expected not to find test.2.txt, got ", result)
    }
    
    teardown()
}


/*func TestMain(m *testing.M) { 

    retCode := m.Run()

    // your func
    teardown()

    // call with result of m.Run()
    os.Exit(retCode)
}*/
