import React, { useState } from "react";
import ExecuteButton from "../components/ExecuteButton";
import Output from "../components/Output";
import CommandService from "../services/CommandService";

const Execution = () => {
  const [input, setInput] = useState("");
  const [output, setOutput] = useState("# Aquí se verán todos los mensajes de la ejecución");

  // Función para ejecutar los comandos
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

  return (
    <div className="container mt-5">
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
  );
};

export default Execution;
