import React, { useState, useEffect } from 'react';

import './RoomGrid.css';


const RoomGrid = () => {
    const [data, setData] = useState([]);

    useEffect(() => {
        const fetchData = async () => {
            try {
                const response = await fetch('http://12.12.12.10:8080/rooms');
                const result = await response.json();
                console.log(result)
                setData(result);
            } catch (error) {
                console.error('Error fetching data:', error);
            }
        };

        fetchData();
    }, []);

    return (
        <div className="container">
            <div className='title'>
                <p className>Rooms:</p>
            </div>
            <div className='RoomGrid'>
                {Object.keys(data).map((roomKey) => (
                    <div onClick={() => {
                        window.location.href = `/rooms/${roomKey}`;
                        this.props.navigation.navigate('Details', {
                            roomName: roomKey,
                        });
                    }
                    } key={roomKey} className="grid-item">
                        <div className="rounded-rectangle">
                            <p>Room Name: {roomKey}</p>
                            <p>Users: {data[roomKey].ClientCount} / {data[roomKey].Capacity}</p>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};

export default RoomGrid;
