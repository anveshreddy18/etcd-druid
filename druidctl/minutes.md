->Add examples for how to use the kubectl plugin with the specific subcommand.

->test the logging on windows machines as well

-> Don't show banner always, show it only when showing usage.

-> the cli folder is not needed, just have a cmd folder and group related stuff together.
->error streams for testing, unit & integration tests need proper error streams.
->yaml/json output, specific to each command. Printers: https://github.com/kubernetes/cli-runtime/tree/master/pkg/printers, https://github.com/kubernetes/kubectl/tree/master/pkg/cmd/get
->make pkgs inside the cmd folder for each command, group all related information together. Get rid of the cli folder. All types should ideally be along with the command, except for the common command context
->the output.Service is not actually a service but just a logger, rename it. 
->easily consumable output from the cli for each command, ideally in a single screen that keeps updating itself. This command specific output rendering should be specific to this particular command, and they all can use the low-level output component from the charm implementation of the logger. TBH, the charm pkg currently only does logging, but if we want more kinds of output rendering, we need better low-level components, relook at this.
-> DEEPLY READ CODE FROM THE EXISTING KUBECTL CODE, RECENT GOOD KUBECTL PLUGINS, see what's good & bad, form your own opinions around them and implement them in your code. Reading code is not bad, infact that's the best way of becoming a better developer.
-> Try making the resume-reconcile & suspend-reconcile as subcommands for the reconcile command itself.


-> Prompt the user when they use -A to get confirmation
-> check if resource resolver or similar functionality is provided by k8s libraries and use that instead.
-> kubectl druid reconcile ns1/etcd1,ns2/* --wait-till-ready
-> when doing the above, few things to take care of.
-> --watch flag to keep on watching, just like `kubectl get pods -w`. watch is basically --wait-till-ready + --timeout=infinite.
-> By default, instead of printing the logs, print rows and columns like `kubectl get pods -w` where every 10 seconds, all the selected etcds with their status gets printed, separate the each iteration with some line separators, and let say in a specific iteration, some of the etcds get finished, then we mark their status lines in green & not keep on updating these etcds in the next iteration. In the next iteration, only update those that are still not ready.

Example output for 4 etcds:

NAME            RECONCILED              UPDATED             READY           TIMESINCE       DONE
ns1/etcd1       True                    False               False           30s             X
ns2/etcd2       True                    True                False           1m20s           X
ns3/etcd3       False                   False               False           1s              X
ns4/etcd4       True                    True                True            2m              tick

----------------------------------------------------------------------------------------

NAME            RECONCILED              UPDATED             READY           TIMESINCE       DONE
ns1/etcd1       True                    True                False           40s             X
ns2/etcd2       True                    True                True            1m30s           tick
ns3/etcd3       True                    False               False           11s             X

----------------------------------------------------------------------------------------


Solution: May be a map to store the states? or a queue that records the events as they come by? the latter looks unnecessary & error prone.