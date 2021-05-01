# E-SAFE

50.041 Distributed Systems Project

> A Distributed Storage Solution that allows for Identity and Access Management (IAM) and secure storage of Secrets while distributing security liability.​

## How to run?

To start as Locksmith Server:

```
go run cmd/e-safe/main.go -locksmith
```

To start as a Node:

```
go run cmd/e-safe/main.go -node
```

To start the front-end web application:

```
cd client
npm install
npm run serve
```

Visit the website at: `localhost:8080`

## Demo

[Video Demo](https://www.youtube.com/watch?v=NAAIVyq9gcU&feature=youtu.be)


---


<center>
    <img src="https://i.imgur.com/OBiQN9X.png" />
</center>
===

<center>
<h1>
<a href="https://youtu.be/NAAIVyq9gcU">Video Demo</a>
</h1>
</center>

<center>
<h5>
<a href="https://github.com/xmliszt/e-safe">Project GitHub Link</a>
</h5>
</center>

<center>
<hr>
<div>Yuxuan 1003607</div>
<div>Jia Yi 1003696</div>
<div>Johnson 1003651</div>
<div>Dev 1003375</div>
<div>Zewen 1003623</div>
</center>
**Team**

---

[TOC]

---

# Introduction and Problem Formulation
Our group aims to design and implement a Distributed Storage Solution that allows for Identity and Access Management (IAM) and secure storage of Secrets while distributing security liability. We aim to mimic DynamoDB's storage system in order to create this secret storage system.

# Overall System Architecture

![](https://i.imgur.com/o3of8um.jpg)


E-Safe distributed system design is as shown above. Nodes are arranged in a ring structure around the Locksmith server. All nodes are communicating with one another via Remote Procedure Call (RPC). Each node has a unique RPC address to constantly listen to internal RPC requests from other nodes.

The Locksmith server is responsible for periodically checking the heartbeat (i.e. whether the node is alive) of each node available, monitoring the coordinator status. If the current coordinator node is down, the Locksmith server will assign coordinator to the next alive node with the highest node ID. 

The Coordinator node is responsible for activating the router within to listen for incoming clients' requests from the end-users. The Controller layer will validate the requests and handle them by calling the service layer. Service layer will then call the Repository layer to make changes or retrieve information from the local file storage. The Service layer will also send RPC calls to other nodes in times of data replication and re-distribution.

Each node's start-up process will include sending RPC requests to the Locksmith to get the latest copy of the shared resources (namely the Heartbeat Table, the RPC addresses, as well as the Virtual Node Structure) and broadcast to every other node. This is to ensure that the shared resources in each node are consistent throughout the system.

# Design

## Terminology

- **Locksmith server**: The central server that checks aliveness of other nodes and controls coordinator assignment.
- **Coordinator**: The special node appointed to listen to clients' requests from Front-end. It is also responsible for hashing the request and computing the replication locations.
- **Node**: A single node that performs main functionality of the system, such as getting secrets, creating secrets, updating secrets, login users, registering users. Each node has a router component but only the coordinator node will activate the router. Each node is identified by its unique ID, starting from 1.
- **Heartbeat table**: The table that shows the aliveness of each node.
- **RPC addresses map**: A mapping of node ID to its respective RPC listener address. Every node's address has the host of `localhost` and the port of `5000 + nodeID`. Locksmith's address is `localhost:5000`. Node ID minimum is 1.
- **Shared Resources**: The shared resources refer to any local data in each node that is suppoded to be consistent in the system. Though they are local, they appear to be *shared* as every node has the excat same values for these data.
- **Owner Node**: The node to which the original data is stored. Owner node is responsible for relaying the replications to the subsequent nodes to store the replicas.
- **Replica Node**: The node to which the replica of the original data is stored.

## Features

### User Features
- **Register user**: User is able to register for new user account, username must be unqiue.
![](https://i.imgur.com/mWqOUME.png)

- **Login user**: User is able to login using his/her username and password.
![](https://i.imgur.com/iH4BENd.png)

- **Create a secret**: User is able to create a new secret - keying in the alias/description of the secret and the value of the secret.

![](https://i.imgur.com/Elxruje.png)

- **View accessible secrets under the role**: User is able to view all the secrets that are accessible by him/her, based on his/her role. Role ranges from 1 to 5, with `role=1` being the role with highest accessibility (i.e. super admin)

![](https://i.imgur.com/gCn8blb.png)

- **Edit a secret**: User is able to edit a secret to a new value

![](https://i.imgur.com/pROglkw.png)

- **Delete a secret**: User is able to delete a existing secret

![](https://i.imgur.com/r9wOsL4.png)

- **Monitor node status**: ***(Demo)*** User is able to monitor the nodes status. Tables showing all virtual nodes and physical nodes status and locations, together with a Doughnut chart will be shown in the monitoring dashboard.
![](https://i.imgur.com/F59kVUm.png)



### Technical Features
- **Fault Detection**: The Locksmith server is able to detect node failure via heartbeat checking and immediately modify shared resources, updating all alive nodes about the changes, and assigning a new coordinator if the dead node happens to be a coordinator.
- **Consistent Hashing**: Each virtual node's name and all incoming data's keys are hashed using the same hashing function. With the ring structure, it makes data re-distribution during system scaling have a smaller overhead.
- **Virtual Nodes**: Each node has a number of virtual locations in the ring structure to achieve *load balancing*.
- **Fault Tolerance**: When a node is down, the request from the client will still be served continuously without the need for retrying, thanks to our data replication.
- **Node Recovery**: A dead node with lost data is able to revive and recover the lost data from other nodes. The Locksmith server is able to handle the recovery and update all shared resources accordingly.
- **Scalability**: Our system is able to scale up and down easily by adding or removing the node, without the worries of lost data or missed clients' requests.

## Assumptions
- **No Byzantine node**: Our group has opted to not implement interventions for malicious, or bayzentine, nodes.
- **Non-volatile storage**: Although Dynamo DB, our reference architecture, tends to implement storage of data onto non-volatile storage that allows them to spin up a new node and ensure data persistency through crashes without the need to search for replicas during recovery - we have opted not to do this. This is partly due to the fact that even with non-volatile storage, a reassurance check should still take place to verify data persistance and correctness. Therefore, with the assumption that a node dying looses all its data, we were able to implement an algorithm that not only recovers the data on node recovery but also redistributes the data as the node position is rehashed on recovery.
- **Locksmith server is fault-proof**: Our locksmith server helps update the heartbeat table and we have assumed that it is fault-proof and therefore cannot die.

## Consistency
Due to the lack of concurrent operations that need to be handled by the system, we do not requeire a specific protocol to handle consistency.

Our system does however guarantee that whatever the client has read/written will be reflected in the database at all times. This is because the router that handles receiving the requests makes sure to queue all requests. This means that there is some form of total order at all times.

We also guarantee data consistency because of our replication algorithm. As long the number of nodes down is leser than or equal to the replication factor, there will always be a replica of the data always available for the client to access. 


## Scalability
### Load Balancing
Our group implements consistent hashing to create **virtual nodes**. Each physical node is hashed into multiple virtual nodes (can be set in the `config.yaml` file). The concept of virtual node is similar to that of DynamoDB, where it helps with load distribution. An example of a physical node ring structure and how it looks like after implementing the concept of virtual nodes can be seen as follows:
![](https://i.imgur.com/DpMSP6W.png)

Virtual nodes allows for data to be redistributed more evenly across all available nodes. As seen from the diagram above, when only physical nodes are implemented, there might be a higher chance where the data is stored in the same node, which makes the burden on a node high. However, with the the implementation of virtual nodes, it spreads out the structure of nodes, where each physical node is responsible for a more random range of data to be stored.

### Addition and Deletion nodes
We implement the concept of **consistent hashing**, where each virtual node is hashed into each individual location based on a unique hash function, where it hashes the virtual node ring string three times to create more randomness, and output a unique 32-bit integer. Due to the uniqueness of hashing, every unique input will result into the a unique output, therefore, no overlapped locations will be created.

When a new node is spawned, with virtual nodes created, it will be inserted into the virtual node ring structure after passing through our unique hash function. It will then grab its secrets from its neighbours and store it in itself. As the data is stored in a clockwise manner, it will get its data from the subsequent node, claiming itself as the owner node of the data. On top of that, it will query the previous nodes's replicated data to store in itself as well.

![](https://i.imgur.com/KRWTZF4.png)


When a node crashes, we assume that the data is wiped out. The node structure will change accordingly as well.

## Fault Tolerance
### Failure Detection
- Heartbeat Table
    - The locksmith server will ping each member of the ring structure to check for liveness. In the scenario a node has been found to be dead (heartbeat == false), the locksmith server will trigger a recovery action based on the node that has died. If it was a Coordinator, the locksmith server will assign the next highest ID node as the Coordinator and the handover process will ensue. In all cases, data redistribution between the remaining nodes will take place. When the node recovers, the Node ID will be hashed again as it is inserted into the ring and data is redistributed again.

### Replication strategy
Given a repliation factor ```n```, our system ensures that we have `n` replicas in the system at all times. A pictorial representation of the replication strategy can be seen below.

![](https://i.imgur.com/f8t1slX.png)

Our replication strategy occurs in 2 phases. The first phase is shown in the picture by the green and blue arrows numbered 1 to 4. When the coordinator receives a request to write data into the system, it first finds the owner virtual node that is responsible for the data. In this case, this will be node 1-3. The coordinator then sends the data to the owner node, along with a list of all the nodes that can store the replica of the data being sent, shown by arrow 1 in the picture. 

This list is generated according to two conditions. The replica virtual node must be a succeeding virtual node from the owner virtual node in a clockwise manner, and the physical node of the replica node must not already exist in the list. For example, in the picture above, we skip node 5-1 since we are already storing the replica in the physical node 5 via vitual node 5-3.

Once the owner virtual node receives and stores the data from the coordinator, it sends the data to the first node in the generated list. It then requires this subsequent node to store the replica data and send an acknowledge message back to the owner virtual node. This ensures that at least replica is guaranteed to be stored in system at all times. This is shown by arrow 2 and 3 in the picture.

Once the owner virtual node receives the acknowledge message from the first node in the list, it sends an acknowledgement message back to the coordinator to conclude the operation. This is shown by arrow 4.

The second phase of the strategy starts here. The virtual node with the replica will then itself look at the next virtual node in the list and send the data over to it for storage. This subsequent node will do the same until all the nodes in the generated list have been traversed. This is shown by arrows 5 and 6 in the picture. We decided not to implement acknowledgement messages for all `n` replicas to trade-off between the time the client has to wait to receive an acknowledgement of a successful write to the system and the guaranteed creation of the replica of data.

### Tolerance to dead node
- Supports up to (Replication Factor) nodes dead.
    - To make sure the system has at least one replica of data, our system allows a maximum of the replication factor (3) number of nodes to die. Under this condition, for the worst case, the next two nodes of the owner node die. The list of the next three nodes which were generated by the coordinator will only contain only one node location, and the owner node will only send strict consistency with the replication factor of 1 to that location. After the third node store the replica in its storage, it will send acknowledge message to the coordinator. Under this condition, at least one replica is stored and can be found when required.



- When node down, replication continues as per normal
    - When the number of dead nodes is smaller than the replication factor, the coordinator will generate a list of node locations that contain the replication factor (3) less the number of the dead nodes number ```(RF - No. of dead nodes)``` of nodes. This will then be sent to the owner node which will start the replication process based on the list received. An example of one physical being down is shown in the picture below

![](https://i.imgur.com/uDoGpfA.png)

- Service is not affected when a node is down
    - Since the system guarantees that there is at least one additional replica stored in some neighboring node. Under this condition, even if the owner node is dead, the system will still be able to get the replica of the information form the neighboring node before sending it to the client. 

### Recovery
There isn't a distinct recovery algorithm that was implemented for our system. This is because the recovery algorithm is akin to adding a new node to the system.

When a node dies, it is removed from the ring structure and the heartbeat table is updated to reflect that it has died. If it wakes up again, the Locksmith treats this node as a new node that is being added to the ring structure, and runs the 'add-node' algorithm to add the revived node back into the ring structure.

In the event the coordinator node dies, our election algorithm reassigns the coordinator responsibilities and status to the next node that has the highest PID. When the previous coordinator node reawakens, the election algorithm is run again and the revived node is given back the status and responsibilities as coordinator.

## Limitations
- Fixed storage overall
    - While we were able to add more nodes into our distributed storage solution, this only helped us in balancing load of incoming messages. The overall capacity for key-value pairs that we were able to store was limited by our Hash Function at ```2^32```. While this is a large number, it is still a hard limit on our entire system as trying to store the ```2^32 + 1```th key-value pair will definitely result in a clash.
- Number of nodes must be greater than the Replication Factor
    - The number of nodes that our distributed system requires to be operating effectively in the said manner described above requires it to have more physical nodes than the number of the Replication Factor ```(No. of Physical Nodes > RF)```
- No auto-scaling 
    - We have not implemented a system that is able to automatically guage the load of each individual node and trigger a scale-up, or scale-down, event. This currently takes place manually when we introduce a new node to the distributed store or kill a node manually.

# Conclusion

Through project e-safe, we have been able to implement a distributed storage solution for secrets with Identity and Access Management (IAM). This allows teams and enterprises to ensure IAM is maintained while their secrets are stored in a distributed system that allows for a reduced security liability. 

We have made this system reliable (ensuring correctness), fault-tolerant and aware (fault-detection). This system's original inspiration was Dynamo DB and some of the protocols and design principles seen here are a direct influence of Dynamo's design. 

With e-safe, we believe we have been able to create a secure and reliable storage for enterprises and teams to store their e-secrets.  

# Appendix


**Base URL**
`http://localhost:8081`

| Name | Endpoint | Method | Available Status |
| -------- | -------- | -------- | ---------- |
| Login    | `/login` |  POST    | 200, 401, 404, 500|
| Register | `/register`| POST | 200, 400, 409, 500 |
| Get a secret | `/api/v1/secret?alias={}` | GET | 200, 400, 401, 500 |
| Create/Update a secret | `/api/v1/secret` | PUT | 200, 400, 401, 500 |
| Delete a secret | `/api/v1/secret?alias={}` | DELETE | 200, 400, 401, 500 |
| Get all secrets | `/api/v1/secrets` | GET | 200, 400, 401, 500 |
