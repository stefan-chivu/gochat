import React, { Component } from "react";
import Header from '../../components/Header/Header';
import ChatHistory from '../../components/ChatHistory/ChatHistory';
import ChatInput from '../../components/ChatInput/ChatInput';
import { sendMsg, connectRoom, getRoomMessages } from '../../api/room'

class RoomPage extends Component {
    constructor(props) {
        super(props);
        this.state = {
            chatHistory: getRoomMessages(),
            username: "",
            isPromptCompleted: false,
        }
    }

    showPrompt = () => {
        const username = window.prompt('Username:');
        console.log("username: " + username)
        this.setState({ username: username, isPromptCompleted: true }, () => {
            console.log("Connecting as " + username)
            connectRoom((msg) => {
                console.log("New Message")
                this.setState(prevState => ({
                    chatHistory: [...this.state.chatHistory, msg]
                }))
                // console.log(this.state);
            }, "Global", username);
        });
    };

    componentDidMount() {
        this.showPrompt();
    }

    send(event) {
        if (event.keyCode === 13) {
            sendMsg(event.target.value);
            event.target.value = "";
        }
    }

    render() {
        if (!this.state.isPromptCompleted) {
            // Render nothing or a loading indicator while waiting for the prompt
            return null;
        }
        return (
            <div className='RoomPage'>
                <Header />
                <h2>Room Global</h2>
                <ChatHistory chatHistory={this.state.chatHistory} />
                <ChatInput send={this.send} />
            </div>
        );
    };
}

export default RoomPage;