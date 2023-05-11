import 'package:flutter/material.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'package:web_socket_channel/io.dart';
import 'dart:convert';

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
  List<String> messages = [];

  @override
  void initState() {
    super.initState();
    channel = WebSocketChannel.connect(Uri.parse('ws://127.0.0.1:8080/chat'));
    channel.stream.listen((dynamic message) {
      // Handle incoming messages from the server
      final jsonData = jsonDecode(message);
      final List<dynamic> messageData = jsonData as List<dynamic>;
      final List<String> receivedMessages =
          messageData.map<String>((dynamic data) {
        final Map<String, dynamic> messageMap = data as Map<String, dynamic>;
        final String username = messageMap['username'] as String;
        final String text = messageMap['text'] as String;
        final String timestamp = messageMap['timestamp'] as String;
        return '[$timestamp] $username: $text';
      }).toList();
      setState(() {
        messages = receivedMessages;
      });
    });
  }

  void sendMessage(String message) {
    final Map<String, dynamic> data = {
      'username': 'User',
      'text': message,
    };
    final jsonData = jsonEncode(data);
    channel.sink.add(jsonData);
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
          Expanded(
            child: ListView.builder(
              itemCount: messages.length,
              itemBuilder: (context, index) {
                return ListTile(
                  title: Text(messages[index]),
                );
              },
            ),
          ),
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
        ],
      ),
    );
  }
}
