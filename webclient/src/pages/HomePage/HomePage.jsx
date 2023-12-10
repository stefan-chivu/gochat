import React, { Component } from "react";
import RoomGrid from "../../components/RoomGrid/RoomGrid"
import Header from '../../components/Header/Header';

class HomePage extends Component {
    render() {
        return (
            <div className='Home'>
                <Header />
                <h2>Home</h2>
                <RoomGrid />
            </div>
        );
    };
}

export default HomePage;