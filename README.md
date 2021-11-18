
## Server Types

- [Unary](#unary-type)
- [Stream Server](#stream-server)
- [Stream Client](#stream-client)
- [Bi-directional Stream](#bi-directional-stream)


With gRPC you can choose between four different communication patterns. Unary which is open a connection, send the request, and get the response. Stream server where the connection is opened and a request is sent, but this time the server continues to send response messages till it determines that it's finished. The client keeps listening till the end of the messages. Stream client where the opposite of the stream server happens. The client opens the connection and sends a stream of request messages . When the client finishes sending requests the server responds and the connection is closed. Last there is the dual stream setup. This one is a bit different. The last two types the unary side of the conversation needed to wait for the end of the stream, but in the dual type requests and responses can travel back and forth at during the connection.

###### Useful Links
- [Basic server/client setups for all types](https://grpc.io/docs/languages/go/basics/)
- [Google protobuf docs](https://developers.google.com/protocol-buffers/docs/overview)
- [Grpc documentation](https://pkg.go.dev/google.golang.org/grpc#Server)
- [Case study using gRPC streams](https://ably.com/blog/grpc-stream-performance)

##### Unary {#unary-type}
I think that we're all pretty familiar with the unary type of gRPC server pattern. None the less we'll go through the server setup for the Hello Service that I've setup here.

In ```unary/internal/hellosvc``` is the ```hellosvc``` package that fulfills the hello service interface to register as the concrete type to the gRPC server on ```line 19 main.go``` 

The server has reflection enabled on ```line 26 main.go``` for easier service discovery. 

With the unary hello service running using a couple grpcurl commands we can find out a lot about the hello service.

    grpcurl --plaintext localhost:50051 list
returns

    grpc.reflection.v1alpha.ServerReflection
    hello.v1.HelloService

then
    
    grpcurl --plaintext localhost:50051 list hello.v1.HelloService

returns

    hello.v1.HelloService.Hello

then

    grpcurl --plaintext localhost:50051 describe hello.v1.HelloService

returns

```
hello.v1.HelloService is a service:
service HelloService {
  rpc Hello ( .hello.v1.HelloRequest ) returns ( .hello.v1.HelloResponse );
}
```  

then

    grpcurl --plaintext localhost:50051 describe hello.v1.HelloService.Hello

returns

```
hello.v1.HelloService.Hello is a method:
rpc Hello ( .hello.v1.HelloRequest ) returns ( .hello.v1.HelloResponse );
```

then 

    grpcurl --plaintext localhost:50051 describe hello.v1.HelloRequest

returns

```
hello.v1.HelloRequest is a message:
message HelloRequest {
  string name = 1;
}
```

and

    grpcurl --plaintext localhost:50051 describe hello.v1.HelloResponse

returns

```
hello.v1.HelloResponse is a message:
message HelloResponse {
  string greeting = 1;
}
```

Finally, sending 

    grpcurl --plaintext -d '{"name":"Frank"}'  localhost:50051  hello.v1.HelloService/Hello

returns

```
{
  "greeting": "Hello Frank"
}
```


##### Stream Server {#stream-server}
If we take a look at the proto for the Name Service in ```stream_server/pkg/namesvc``` the proto looks the same as most proto files, except in one case. THe returned ```GetNamesResponse``` is preceeded by the ```stream``` key word. This tells the proto compiler that we want to return a stream of this type of message to the client. This is all we need to do to create the stream server gRPC packages. The compiler will do the rest.

For the stream server I've set up a client as well as a server for the Name Service. The stream server set up in main looks the same as usual. In the nameservice package which is the implementation of the Name Service, there is the GetNames method. This looks a bit different. The method takes a request value and a stream server for it's second argument. It's used to send as many response messages as needed. Here we send a random number of name responses with the ```stream.Send()``` method. 

Most of the complexity of the stream function is taken care of by the proto compiler for go. We just get to use this stream server given to the GetNames method by the ```nameservice``` package. 

The client calling the stream request gets a stream client that the response messages can be harvested from till we got the go idiomatic ```io.EOF``` error telling us we've reached the end.

I'll open a terminal and ```cd``` into the stream server server directory, and start the server ```go run main.go```.

In the terminal I see the info log that the server is running. ```INFO[0000] Name Server Running```

Next I'll open another terminal and ```cd``` into the client directory, and start the client. ```go run main.go```

The client logs a list of names sent from the server. 
```
Got Names:
Abbigail Leannon
Lewis Huels
Stanton Dicki
...
```

The server logs an info log telling us that the names are sending ```INFO[0133] Sending Names``` and continues to wait for requests.

We could also use a gRPC cli tool to make the request to the server. This time lets use the gRPC Evans cli tool.

I'll open Evans in another terminal using the command ```evans -r repl```, and Evans opens with a prompt indicating that it's aware of Name Service ```namesservice.NamesService@127.0.0.1:50051>```. The ```-r``` flag passed to the ```evans``` command says that we expect the server to have reflection enabled, and the name server does. 

At the prompt we can ```show service``` and get the NameService methods available and the types that the methods require and return. For the NameService there is only the ```GetNames``` method that takes the ```GetNamesRequest``` and returns a ```GetNamesResponse```

Like with grpcurl we can get a description of the messages by using the ```desc``` command. ```desc GetNamesRequest``` shows us a table with nothing in it because the ```GetNames``` method takes an empty request type ```message GetNamesRequest {};```. The response message ```GetNamesResponse``` description ```desc GetNamesResponse``` shows that the message has one ```name``` attribute of type string and it isn't repeated ( in the message at least :) ). 

I'll call the ```GetNames``` method with ```call GetNames```. Since ```GetNames``` ```GetNamesRequest``` is empty no additional information is required to call the ```GetNames``` method. Similar to the cliern log we get multiple ```GetNamesResponse``` messages and Evans lists them as they come in.

```
{
  "name": "Neha Schinner"
}
{
  "name": "Princess Dibbert"
}
{
  "name": "Anita Smitham"
}
{
  "name": "Deanna Stanton"
}
...
```

##### Stream Client {#stream-client}
Looking at the proto for the Hub Updater Service the ```UpdateStatusEvents``` method takes a ```stream UpdateStatusEventRequest```. Again the ```stream``` key word precedes the message type , but in this case it's the method argument that is preceeded with ```stream```. This tells the proto compiler that we want to send multiple request message types to the client to process before sending the response.

The ```hubservice``` package in ```stream_client/internal/hubservice``` is the implementation of the ```HubUpdaterService``` defined in the proto.

The ```UpdateStatusEvent``` method takes only a stream server since the stream is now a stream of request messages. 

The stream is read in a similar maner to the client reading the stream in the stream server example, but here after the final request message comes in, indicated by the ```io.EOF``` error. 

When the server gets the end of requests it can send the response via the stream server ```SendAndClose``` server method. In this case the server responds with the list eventIds that were sent.

I'll open a terminal and ```cd``` into the stream server server directory, and start the server ```go run main.go```.

In the terminal I see the info log that the server is running. ```INFO[0000] Hub Update Server Running```

Next I'll open another terminal and ```cd``` into the client directory, and start the client. ```go run main.go```

The client sends a group of events and the server responds with the eventIds. A log from the client displays the eventIds.

```
Updated events [fa8a9140-eff7-488f-afbb-cf5f52d659c9 42306fe5-58ee-4300-9127-8ae282ee2430 7445e42d-73d3-44de-accc-876d727dfabf 93fdd808-4b52-4c77-b0d4-d28ec96c4e6c ...]
```

##### Bi-directional Stream {#bi-directional-stream}
If we look in the dual stream proto in ```dual_stream/pkg/talkingservice``` you'll probably guess what we'd do to set up bi-directional communication between the client and server. The ```TalkingService``` ```Chat``` methods ```ChatRequest``` and return ```ChatResponse``` are preceded by the ```stream``` keyword telling the proto compiler to generate both a stream client and server.

The ```Chat``` method in the ```talkingservice``` package takes only a stream server. In the ```TalkingService``` setup I've chosen to respond to every request sent by a client. On invocation of the ```Chat``` method a loop is initiated calling ```talkStream.Recv()```. ```Chat``` looks for the end of messages to terminate the lool, but it could wait for the end of messages to respond, or respond after the first and no other message. 

The client setup is a bit different. After the stream client is initiated ```clientStream, err := chatClient.Chat(context.Background())``` the client starts listening for server responses in a loop running in a go routine. The chat receiver go routine watches for the end of messages indicated by ```io.EOF``` and when it gets it, first it closes the ```stopRecieve``` channel that is keeping ```main``` from closing before all the responses come in, then it closes the go routine. We've already seen the stream ```Recv``` method in the previous two examples. If not ```io.EOF``` it processes the response and go back for another.

After the client starts listening it can begin sending the request messages on the ```clientStream``` with the stream ```Send``` method. 

Again I'll open a terminal and ```cd``` into the stream server server directory, and start the server ```go run main.go```.

In the terminal I see the info log that the server is running. ```INFO[0000] Chat Server Running```

Next I'll open another terminal and ```cd``` into the client directory, and start the client. ```go run main.go```

The server logs the chat requests and the generated response. This isn't hard because the server recieves the request and sends a message back in response, but the client needs to use the chatIds to relate chat responses to chat requests. The client also logs the chats related by that chatId so that a user gets the chat response ment for the user. 

Server log:

```
User: Frank Johnson
Chats
User chat: Yo server. Greetings from Frank Johnson. What you got?
	Response: Consequatur voluptatem accusantium perferendis sit aut.
User chat: Yo server. Greetings from Frank Johnson. What you got?
	Response: Perferendis consequatur voluptatem sit aut accusantium.
User chat: Yo server. Greetings from Frank Johnson. What you got?
	Response: Perferendis accusantium sit aut voluptatem consequatur.
User chat: Yo server. Greetings from Frank Johnson. What you got?
	Response: Voluptatem sit consequatur accusantium perferendis aut.
```

Client log:

```
Chat Request: Yo server. Greetings from Frank Johnson. What you got?
Chat Response: Consequatur voluptatem accusantium perferendis sit aut.
Chat Request: Yo server. Greetings from Bill Williams. What you got?
Chat Response: Consequatur aut accusantium sit perferendis voluptatem.
Chat Request: Yo server. Greetings from Maybel Baker. What you got?
Chat Response: Sit voluptatem accusantium aut consequatur perferendis.
Chat Request: Yo server. Greetings from Frank Johnson. What you got?
Chat Response: Perferendis consequatur voluptatem sit aut accusantium.
Chat Request: Yo server. Greetings from Bill Williams. What you got?
Chat Response: Voluptatem aut consequatur perferendis accusantium sit.
...
```
