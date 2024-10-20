import React from "react";
import { BrowserRouter as Router, Route, Routes } from "react-router-dom";
import Navbar from "./components/Navbar";
import Execution from "./pages/Execution";
import Login from './pages/Login'; 
import Visualizador from './pages/Visualizador'; 

const App = () => {
  return (
    <Router>
      <Navbar />
      <Routes> 
        <Route path="/execution" element={<Execution />} />
        <Route path="/login" element={<Login/>} />
        <Route path="/visualizador" element={<Visualizador />} /> 
      </Routes>
    </Router>
  );
};

export default App;
