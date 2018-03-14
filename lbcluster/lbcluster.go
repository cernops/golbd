package lbcluster

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
)

const WorstValue int = 99999
const TIMEOUT int = 10
const OID string = ".1.3.6.1.4.1.96.255.1"

type LBCluster struct {
	Cluster_name            string
	Loadbalancing_username  string
	Loadbalancing_password  string
	Host_metric_table       map[string]int
	Parameters              Params
	Time_of_last_evaluation time.Time
	Current_best_hosts      []string
	Previous_best_hosts     []string
	Previous_best_hosts_dns []string
	Statistics_filename     string
	Per_cluster_filename    string
	Current_index           int
	Slog                    *Log
}

type Params struct {
	Behaviour        string
	Best_hosts       int
	External         bool
	Metric           string
	Polling_interval int
	Statistics       string
}

// Shuffle pseudo-randomizes the order of elements.
// n is the number of elements. Shuffle panics if n < 0.
// swap swaps the elements with indexes i and j.
func Shuffle(n int, swap func(i, j int)) {
	if n < 0 {
		panic("invalid argument to Shuffle")
	}

	// Fisher-Yates shuffle: https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle
	// Shuffle really ought not be called with n that doesn't fit in 32 bits.
	// Not only will it take a very long time, but with 2³¹! possible permutations,
	// there's no way that any PRNG can have a big enough internal state to
	// generate even a minuscule percentage of the possible permutations.
	// Nevertheless, the right API signature accepts an int n, so handle it as best we can.
	i := n - 1
	for ; i > 1<<31-1-1; i-- {
		j := int(rand.Int63n(int64(i + 1)))
		swap(i, j)
	}
	for ; i > 0; i-- {
		j := int(rand.Int31n(int32(i + 1)))
		swap(i, j)
	}
}

type Pair struct {
	Key   string
	Value int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

/* I don't think we need this anymore. We can create the statistics based on the information from timber

func (self *LBCluster) initialize_statistics() error {
	hostlist := make([]string, len(self.Host_metric_table))
	i := 0
	for k := range self.Host_metric_table {
		hostlist[i] = k
		i++
	}
	sort.Strings(hostlist)
	f, err := os.OpenFile(self.Statistics_filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0640)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "date\t\ttime\t\t")
	for i := range self.Current_best_hosts {
		_, err = fmt.Fprintf(f, "tcbh_%d\t\ttcbh_%d_metric\t", i+1, i+1)

	}
	if self.Parameters.Statistics == "long" {
		for _, host := range hostlist {
			_, err = fmt.Fprintf(f, "%s\t", host)
		}
	}
	_, err = fmt.Fprintf(f, "\n")
	return err
}

func (self *LBCluster) Create_statistics() error {
	var err error
	if self.Parameters.Statistics != "none" {
		fi, err := os.Stat(self.Statistics_filename)
		if os.IsNotExist(err) || (fi.Size() == 0) {
			self.initialize_statistics()
		}
		f, err := os.OpenFile(self.Statistics_filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0640)
		if err != nil {
			return err
		}
		defer f.Close()
		t := time.Now()
		timestamp := fmt.Sprintf("%04d-%02d-%02d\t%02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
		_, err = fmt.Fprintf(f, "%v\t", timestamp)
		for _, host := range self.Current_best_hosts {
			if host != "unknown" {
				if len(host) < 8 {
					_, err = fmt.Fprintf(f, "%v\t\t%v\t\t", host, self.Host_metric_table[host])
				} else {
					_, err = fmt.Fprintf(f, "%v\t%v\t\t", host, self.Host_metric_table[host])
				}
			}
		}
		if self.Parameters.Statistics == "long" {
			hostlist := make([]string, len(self.Host_metric_table))
			i := 0
			for k := range self.Host_metric_table {
				hostlist[i] = k
				i++
			}
			sort.Strings(hostlist)
			metric_value := WorstValue
			for _, host := range hostlist {
				metric_value = self.Host_metric_table[host]
				if metric_value == WorstValue {
					// metric_value = 1000 - metric_value
					metric_value = -9
				}
				if metric_value < 0 {
					metric_value = 0
				}
				_, err = fmt.Fprintf(f, "%v\t", metric_value)
			}
		}
		_, err = fmt.Fprintf(f, "\n")
	}
	return err
}
*/

func (self *LBCluster) Time_to_refresh() bool {
	// self.Write_to_log(fmt.Sprintf("Time_of_last_evaluation = %v now = %v Time_of_last_evaluation + polling_int = %v result = %v Cluster_name = %v\n", self.Time_of_last_evaluation, time.Now(), self.Time_of_last_evaluation.Add(time.Duration(self.Parameters.Polling_interval)*time.Second), self.Time_of_last_evaluation.Add(time.Duration(self.Parameters.Polling_interval)*time.Second).After(time.Now()), self.Cluster_name))
	return self.Time_of_last_evaluation.Add(time.Duration(self.Parameters.Polling_interval) * time.Second).Before(time.Now())
}

func (self *LBCluster) Find_best_hosts() {
	self.Previous_best_hosts = self.Current_best_hosts
	self.evaluate_hosts()
	allMetrics := make(map[string]bool)
	allMetrics["minimum"] = true
	allMetrics["cmsfrontier"] = true
	allMetrics["minino"] = true

	_, ok := allMetrics[self.Parameters.Metric]
	if !ok {
		self.Write_to_log("ERROR", "wrong parameter(metric) in definition of cluster "+self.Parameters.Metric)
		return
	}
	self.apply_metric()
	self.Time_of_last_evaluation = time.Now()
	nodes := strings.Join(self.Current_best_hosts, " ")
	if len(self.Current_best_hosts) == 0 {
		nodes = "NONE"
	}
	self.Write_to_log("INFO", "best hosts are: "+nodes)
}

// Internal functions
/* This is the core of the lbcluster: based on the metrics, select the best hosts */
func (self *LBCluster) apply_metric() {
	self.Write_to_log("INFO", "Got metric = "+self.Parameters.Metric)
	pl := make(PairList, len(self.Host_metric_table))
	i := 0
	for k, v := range self.Host_metric_table {
		pl[i] = Pair{k, v}
		i++
	}
	//Let's shuffle the hosts before sorting them, in case some hosts have the same value
	Shuffle(len(pl), func(i, j int) { pl[i], pl[j] = pl[j], pl[i] })
	sort.Sort(pl)
	self.Write_to_log("DEBUG", fmt.Sprintf("%v", pl))
	var sorted_host_list []string
	var useful_host_list []string
	for _, v := range pl {
		if (v.Value > 0) && (v.Value <= WorstValue) {
			useful_host_list = append(useful_host_list, v.Key)
		}
		sorted_host_list = append(sorted_host_list, v.Key)
	}
	self.Write_to_log("DEBUG", fmt.Sprintf("%v", useful_host_list))
	useful_hosts := len(useful_host_list)
	list_length := len(pl)
	max := self.Parameters.Best_hosts
	if max == -1 {
		max = list_length
	}
	if max > list_length {
		self.Write_to_log("WARNING", fmt.Sprintf("impossible to return %v hosts from the list of %v hosts (%v). Check the configuration of cluster. Returning %v hosts.", max, list_length, sorted_host_list, list_length))
		max = list_length
	}
	if list_length == 0 {
		self.Write_to_log("ERROR", "cluster has no hosts defined ! Check the configuration.")
		self.Current_best_hosts = []string{"unknown"}
	} else if useful_hosts == 0 {
		if self.Parameters.Metric == "minimum" {
			self.Write_to_log("WARNING", fmt.Sprintf("no usable hosts found for cluster! Returning random %v hosts.", max))
			Shuffle(len(sorted_host_list), func(i, j int) {
				sorted_host_list[i], sorted_host_list[j] = sorted_host_list[j], sorted_host_list[i]
			})
			self.Current_best_hosts = sorted_host_list[:max]
		} else if (self.Parameters.Metric == "minino") || (self.Parameters.Metric == "cmsweb") {
			self.Write_to_log("WARNING", "no usable hosts found for cluster! Returning no hosts.")
			self.Current_best_hosts = useful_host_list
		} else if self.Parameters.Metric == "cmsfrontier" {
			self.Write_to_log("WARNING", "no usable hosts found for cluster!, using the previous_best_hosts")
			self.Current_best_hosts = self.Previous_best_hosts
		}
	} else {
		if useful_hosts < max {
			self.Write_to_log("WARNING", fmt.Sprintf("only %v useable hosts found in cluster", useful_hosts))
			max = useful_hosts
		}
		self.Current_best_hosts = useful_host_list[:max]
		self.Slog.Debug(fmt.Sprintf("%v", self.Current_best_hosts))
	}
	return
}
