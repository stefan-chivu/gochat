import React, { Component } from "react";
import "./App.css";
import {
  BrowserRouter as Router,
  Routes,
  Route,
} from "react-router-dom";
import HomePage from "./pages/HomePage";
import RoomPage from "./pages/RoomPage";



class App extends Component {
  constructor(props) {
    super(props);
  }

  render() {
    return (
      <div className="App">
        <Router>
          {/* <Navbar /> */}
          <Routes>
            <Route exact path="/" element={<HomePage />} />
            <Route path="/rooms/:roomName" element={<RoomPage />} />
            {/* <Route path="/about" element={<About />} />
          <Route
            path="/contact"
            element={<Contact />}
          />
          <Route path="/blogs" element={<Blogs />} />
          <Route
            path="/sign-up"
            element={<SignUp />}
          /> */}
          </Routes>
        </Router>
      </div>
    );
  }
}

export default App;