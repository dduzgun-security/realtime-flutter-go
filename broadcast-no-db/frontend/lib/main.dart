import 'package:flutter/material.dart';
import 'package:web_socket_channel/io.dart';
import 'package:web_socket_channel/web_socket_channel.dart';

void main() => runApp(MyApp());

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'WebSocket Demo',
      theme: ThemeData(primarySwatch: Colors.blue),
      home: MyHomePage(),
    );
  }
}

class MyHomePage extends StatefulWidget {
  @override
  _MyHomePageState createState() => _MyHomePageState();
}

class _MyHomePageState extends State<MyHomePage> {
  late final WebSocketChannel channel;
  final TextEditingController _textController = TextEditingController();

  @override
  void initState() {
    super.initState();

    channel = WebSocketChannel.connect(Uri.parse('ws://127.0.0.1:8080/ws'));
  }

  void sendMessage(String message) {
    channel.sink.add(message);
    _textController
        .clear(); // Clear the text input field after sending the message
  }

  @override
  void dispose() {
    channel.sink.close();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('WebSocket Demo')),
      body: Column(
        children: [
          TextField(
            controller: _textController,
            decoration: InputDecoration(
              labelText: 'Message',
            ),
          ),
          ElevatedButton(
            onPressed: () => sendMessage(_textController.text),
            child: Text('Send Message'),
          ),
          StreamBuilder(
            stream: channel.stream,
            builder: (context, snapshot) {
              if (snapshot.hasData) {
                final message = snapshot.data as String;
                return Text('Received: $message');
              } else if (snapshot.hasError) {
                return Text('Error: ${snapshot.error}');
              } else {
                return CircularProgressIndicator();
              }
            },
          ),
        ],
      ),
    );
  }
}
