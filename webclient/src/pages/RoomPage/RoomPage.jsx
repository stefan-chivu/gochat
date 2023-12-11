import React, { Component } from "react";
import Header from '../../components/Header/Header';
import ChatInput from '../../components/ChatInput/ChatInput';
import Sidebar from "../../components/Sidebar/Sidebar";

import { sendMsg, connectRoom } from '../../api/room';

class RoomPage extends Component {
    constructor(props) {
        super(props);
        this.roomName = this.getRoomFromUrl()
        this.state = {
            roomHistory: [],
            username: "",
            isPromptCompleted: false,
        }
        // console.log(this.state)
    }

    getRoomFromUrl() {
        // Get the current pathname (e.g., '/rooms/global')
        const pathname = window.location.pathname;

        // Split the pathname into segments
        const segments = pathname.split('/');

        // Assuming the structure is always /rooms/{roomName},
        // the room name should be at index 2
        const roomName = segments[2];

        return roomName;
    };

    async getRoomMessages(roomName) {
        console.log(`http://12.12.12.10:8080/rooms/${roomName}/messages`)
        const response = await fetch(`http://12.12.12.10:8080/rooms/${roomName}/messages`);
        const result = await response.json();

        return result;
    };


    async showPrompt() {
        const username = window.prompt('Username:');
        console.log("username: " + username)

        const roomHistory = await this.getRoomMessages(this.roomName)

        this.state.roomHistory = roomHistory

        this.setState({ username: username, isPromptCompleted: true }, () => {
            console.log("Connecting as " + username)
            connectRoom((msg) => {
                console.log("New Message")
                this.setState((prevState) => ({
                    roomHistory: [...this.state.roomHistory, msg]
                }))
                console.log(this.state);

            }, "Global", username);
        });
    }

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
                <Sidebar username={this.state.username} />

                <h2>Room Global</h2>
                {this.state.roomHistory.map(msg => {
                    return (
                        <div className="Message">
                            [{msg.timestamp}] {msg.username}: {msg.content}
                        </div>
                    )
                })}
                <ChatInput send={this.send} />
            </div>
        );
    };
}

export default RoomPage;