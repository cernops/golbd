package lbcluster

type LBCluster struct {
	cluster_name            string
	loadbalancing_username  string
	loadbalancing_password  string
	host_metric_table       map[string][]int
	parameters              Params
	time_of_last_evaluation int
	current_best_hosts      []string
	previous_best_hosts     []string
	statistics_filename     string
	per_cluster_filename    string
	current_index           int
}

type Params struct {
	Behaviour        string
	Best_hosts       int
	External         bool
	Metric           string
	Polling_interval int
	Statistics       string
}
