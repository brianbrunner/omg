Strategy For Adding A New Node To The Cluster
=============================================

Two different approaches could work here. The db could use either a continuous hash ring or a large number of buckets. A continous hash ring is easier is some senses, but it requires more *agreement* between the nodes in the cluster. Agreement is tough becuase ensuring all nodes agree on somehting is non-trivial. The other option is to use a large number of buckets, probably something that is a somehwat large power of two (2^14-2^16). Buckets are easier as they can be handed off between machines without agreement from other machines. The only time we have to communicate with machines that are not part of the transaction is when we are looking for a key that belongs in a bucket that is not present in our lookup table or when we want to check whether or not a bucket has achieved the minimum number of replicas the system requires.

Buckets
=======

On Node Add
-----------

A node is added to the cluster by contacting one other node. This node now must broadcast to all other nodes that it can reach that a new node has joined the cluster. Then each node has a small window to submit a list of buckets to the new node. The buckets submitted to the new node can be submitted for two reasons:

1. A particular bucket has not been replicated enough times in the cluster
2. A node is too full and needs to shed buckets to remain active

In the case of 1, we set up the new node as a slave of the node that has the master copy of the bucket in question. In case 2, we opt to transfer a bucket from the server that is not the master for a bucket. If a server is the master of a bucket but needs to shed that bucket in particular, then we first must transfer the master status to a different server. If no other server can become master, the server that is shedding the bucket remains master until the bucket has been fully transferred over.

Gossip Requests
---------------

LISTBUCKETS
returns a multibulkreply whose entries are pairs of bulk replies, the first entry being the bucket id and the second entry being either `M` if the server is the master for that bucket or `S` if the server is a slave for that bucket

REQUESTMOVE bucket_id ip port
asks if the server at ip port can take possesion of the specified bucket. this command will always receive an OKReply. This only acknowledges that the request was received, not that the bucket can be moved.

DUMPBUCKET bucket_id ip port
serializes the specified bucket from the server at ip port to the host that the command is being executed from. When the dump completes, the bucket is deleted from the remote host, and the command is completed.

BUCKETMOVED bucket_id moved_to_ip moved_to_port ip port
Tells the server at ip port that bucket with bucket_id was moved to the server at moved_to_ip and moved_to_port
