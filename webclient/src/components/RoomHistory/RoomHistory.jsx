import React, { Component } from "react";
import "./RoomHistory.scss";
import Message from '../Message/Message';

class RoomHistory extends Component {
    render() {
        console.log(this.props.RoomHistory);
        const messages = this.props.RoomHistory.map(msg => <Message message={msg.data} />);

        return (
            <div className='RoomHistory'>
                <h2>Room</h2>
                {messages}
            </div>
        );
    };
}

export default RoomHistory;