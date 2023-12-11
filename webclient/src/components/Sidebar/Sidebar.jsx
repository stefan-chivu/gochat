import React, { useState, useEffect } from 'react';
import { sendMsg, connect } from '../../api/index';
import './Sidebar.scss'


const Sidebar = ({ username }) => {
    const [data, setData] = useState([]);

    useEffect(() => {
        const showPrompt = () => {
            const username = window.prompt('Username:');
            console.log("username: " + username)
            connect((msg) => {
                console.log("Connecting as " + username)

            }, username);
        };


        const fetchData = async () => {
            try {
                const response = await fetch('http://12.12.12.10:8080/users');
                const result = await response.json();
                console.log(result)
                setData(result);
            } catch (error) {
                console.error('Error fetching data:', error);
            }
        };

        if (username === null || username === "" || username === undefined) {
            showPrompt();
        } else {
            connect((msg) => {
                console.log("Connecting as " + username)

            }, username);
        }
        fetchData();
    }, []);

    return (
        <div className="Sidebar">
            <h2>Active users</h2>
            <ul>
                {Object.keys(data).map((userKey) => (
                    <div onClick={async () => {
                        const roomsResponse = await fetch("http://12.12.12.10:8080/rooms")
                        const roomsResult = await roomsResponse.json();

                        console.log(roomsResult.data)
                        // window.location.href = `/rooms/Private_${username}_${data[userKey]}`;
                    }
                    } key={userKey} className='rounded-rectangle'>
                        <li key={data[userKey]}>{data[userKey]}</li>
                    </div>
                ))
                }
            </ul>

        </div >
    );
};

export default Sidebar;
