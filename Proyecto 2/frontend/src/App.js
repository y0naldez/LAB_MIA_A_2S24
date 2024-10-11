import React from "react";
import { BrowserRouter as Router, Route, Routes } from "react-router-dom";
import Navbar from "./components/Navbar";
import Execution from "./pages/Execution";

const App = () => {
  return (
    <Router>
      <Navbar />
      <Routes> 
        <Route path="/execution" element={<Execution />} />
      </Routes>
    </Router>
  );
};

export default App;
