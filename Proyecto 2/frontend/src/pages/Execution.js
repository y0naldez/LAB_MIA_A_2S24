import React, { useState, useEffect } from "react";
import ExecuteButton from "../components/ExecuteButton";
import Output from "../components/Output";
import CommandService from "../services/CommandService";
import Swal from "sweetalert2";

const Execution = () => {
  const [input, setInput] = useState("");
  const [output, setOutput] = useState("# Aquí se verán todos los mensajes de la ejecución");
  const [loggedUser, setLoggedUser] = useState("");

  useEffect(() => {
    // Leer el usuario logueado desde localStorage
    const user = localStorage.getItem("loggedUser");
    setLoggedUser(user);
  }, []);

  const handleExecute = () => {
    const commands = input.split("\n").filter((cmd) => cmd.trim() !== "");
    commands.forEach((command) => {
      CommandService.parseAndSend(command)
        .then((data) => {
          setOutput((prevOutput) => `${prevOutput}\n${JSON.stringify(data)}`);
        })
        .catch((error) => {
          setOutput((prevOutput) => `${prevOutput}\nError: ${error.message}`);
        });
    });
  };

  const handleLogout = () => {
    Swal.fire("Cerrar sesión", "Funcionalidad aún no implementada", "info");
  };

  return (
    <div className="container mt-5 position-relative">
      {loggedUser && (
        <div className="position-absolute top-0 end-0 m-3">
          <button className="btn btn-success me-2">Usuario {loggedUser}</button>
          <button className="btn btn-danger" onClick={handleLogout}>Cerrar Sesión</button>
        </div>
      )}

      <div className="mt-5 pt-5">
        <h2 className="mb-3">Entrada:</h2>
        <textarea
          className="form-control"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="Introduce aquí los datos de entrada..."
          rows="5"
        ></textarea>
        <ExecuteButton onClick={handleExecute} />
        <h2 className="mt-5 mb-3">Salida:</h2>
        <Output output={output} />
      </div>
    </div>
  );
};

export default Execution;

