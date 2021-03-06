package computing

import (
	"encoding/json"
	"fmt"
	"github.com/timeloveboy/moegraphdb/graphdb"
	"io/ioutil"
	"os"
	"runtime"

	"sort"
	"strconv"

	"sync"
	"time"
)

var Start = false
var Now_vid = 1
var Ids []int
var task chan uint = make(chan uint, 100000)
var result chan map[uint]int = make(chan map[uint]int, runtime.NumCPU())
var Maxfans, Mincount = 100 * 10000, 10

var Result map[uint]int = make(map[uint]int)

func JsonResult() []byte {
	bs, _ := json.Marshal(Result)
	return bs
}

func Mapper(this graphdb.RelateGraph, maxfans, mincount int, myids []int, taskname string) {
	Maxfans = maxfans
	Mincount = mincount
	fmt.Println("start mapping")

	for i := 0; i < 4*runtime.NumCPU(); i++ {
		go re(i, this)
	}

	if len(myids) == 0 {
		Ids = make([]int, 0)
		for v := range this.Users.IterItems() {
			Ids = append(Ids, int(v.Key))
		}

	} else {
		Ids = myids
	}

	sort.Ints(Ids)

	fmt.Println("start jobber " + strconv.Itoa(len(Ids)))
	go func() {
		for _, id := range Ids {
			task <- uint(id)
		}
		fmt.Println("end jobber")
	}()

	fmt.Println("start duce")
	go func(size int) {
		for d := 0; d < size; d++ {
			ducer()
		}
		os.MkdirAll("output", os.ModePerm)
		ioutil.WriteFile("output/"+taskname, JsonResult(), os.ModePerm)
		Start = false
		Now_vid = 1
		result = make(chan map[uint]int)
		Result = make(map[uint]int)
		fmt.Println("end duce")

	}(len(Ids))

}
func re(workid int, this graphdb.RelateGraph) {
	for true {
		vid := <-task
		starttime := time.Now().UnixNano()
		u := this.GetUser(vid)
		vid_likes := make([]uint, 0)
		if u != nil {
			vid_likes = u.Getlikes()
		}
		vid_likes_max1000000 := this.Filterusers_fanscount(vid_likes, Maxfans, 0)
		count_count := this.GetThemCommonFans(vid_likes_max1000000...)
		count_count_10 := graphdb.Filtercount_min(count_count, Mincount, 1<<32)

		result <- count_count_10

		usingtime := time.Now().UnixNano() - starttime
		if usingtime > 1000*1000*100 {
			fmt.Println("workid" + strconv.Itoa(workid) + " is complete " + strconv.Itoa(int(vid)) + "len " + strconv.Itoa(len(count_count_10)) + "all len" + strconv.Itoa(len(result)) + " using milisecond" + fmt.Sprint(usingtime/1000000))
		}
	}
}

var Lock sync.RWMutex

func ducer() {
	c := <-result
	Now_vid++
	Lock.Lock()
	for k, v := range c {
		Result[k] += v
	}
	Lock.Unlock()
}
