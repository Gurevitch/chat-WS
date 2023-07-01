import React, { useState, useEffect, useRef } from "react";
import "./App.css";

const App = () => {
  const [messages, setMessages] = useState([]);
  const [inputValue, setInputValue] = useState("");
  const [authenticated, setAuthenticated] = useState(false);
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const socketRef = useRef(null);

  useEffect(() => {
    // Check if the user is authenticated (you can use an authentication token or session)

    // Simulating authentication status for demonstration
    const isAuthenticated = localStorage.getItem("authenticated") === "true";
    setAuthenticated(isAuthenticated);

    if (isAuthenticated) {
      socketRef.current = new WebSocket("ws://localhost:8000/ws");

      socketRef.current.onopen = () => {
        console.log("WebSocket connected");
      };

      socketRef.current.onmessage = (event) => {
        const message = JSON.parse(event.data);
        setMessages((prevMessages) => [...prevMessages, message]);
      };
    }

    return () => {
      if (isAuthenticated) {
        socketRef.current.close();
      }
    };
  }, []);

  const handleLogin = () => {
    // Perform login request and handle authentication
    // Once authenticated, set the authenticated state to true and store it in localStorage
    // Modify the login logic to suit your API endpoint and response handling
    const loginData = {
      username: username,
      password: password,
    };

    // Send the loginData to your backend API using fetch or Axios
    fetch("http://localhost:8000/login", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(loginData),
    })
      .then((response) => response.json())
      .then((data) => {
        // Check the response from the server to determine if login is successful
        if (data.success) {
          setAuthenticated(true);
          localStorage.setItem("authenticated", "true");
        } else {
          // Handle failed login
          console.log("Login failed");
        }
      })
      .catch((error) => {
        // Handle error
        console.log("Error during login:", error);
      });
  };

  const handleLogout = () => {
    // Perform logout action and handle authentication
    // Once logged out, set the authenticated state to false and remove it from localStorage
    setAuthenticated(false);
    localStorage.removeItem("authenticated");
  };

  const sendMessage = () => {
    if (inputValue.trim() !== "") {
      const message = {
        content: inputValue,
        timestamp: new Date().toUTCString(),
      };
      socketRef.current.send(JSON.stringify(message));
      setInputValue("");
    }
  };

  // Render the login page if not authenticated
  if (!authenticated) {
    return (
      <div className="App">
        <div className="login-container">
          <h1>Login Page</h1>
          <input
            type="text"
            placeholder="Username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
          />
          <input
            type="password"
            placeholder="Password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
          <button onClick={handleLogin}>Login</button>
        </div>
      </div>
    );
  }

  // Render the homepage if authenticated
  return (
    <div className="App">
      <div className="messages-container">
        {messages.map((message, index) => (
          <div key={index} className="message">
            <span className="timestamp">{message.timestamp}</span>
            <span> Message: </span>
            <span className="content">{message.content}</span>
          </div>
        ))}
      </div>
      <div className="input-container">
        <input
          type="text"
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
        />
        <button onClick={sendMessage}>Send</button>
        <button onClick={handleLogout}>Logout</button>
      </div>
    </div>
  );
};

export default App;
