import React, { useState } from "react";
import CommandService from "../services/CommandService";
import Swal from "sweetalert2";

const Login = ({ onLogin }) => {
  const [userId, setUserId] = useState("");
  const [password, setPassword] = useState("");
  const [partitionId, setPartitionId] = useState("");
  const [output, setOutput] = useState("");

  const handleLogin = () => {
    const loginCommand = `login -user="${userId}" -pass="${password}" -id="${partitionId}"`;

    CommandService.parseAndSend(loginCommand)
      .then((response) => {
        console.log("Respuesta completa del servidor:", response);
        
        // Guardar el usuario en localStorage
        localStorage.setItem("loggedUser", userId);
        setOutput(`Login exitoso: ${JSON.stringify(response)}`);
        Swal.fire("Inicio de sesión exitoso", "Bienvenido a la aplicación", "success");

        // Llamar a la función pasada como prop para notificar el login
        if (onLogin) onLogin(userId);
      })
      .catch((error) => {
        console.error("Error completo recibido:", error);
        const errorMessage = error.response?.data?.error || "Error desconocido";
        Swal.fire("Error al iniciar sesión", errorMessage, "error");
        setOutput(`Error: ${errorMessage}`);
      });
  };

  return (
    <div className="container mt-5">
      <h2>Login</h2>
      <div className="mb-3">
        <label>ID Partición</label>
        <input
          type="text"
          className="form-control"
          value={partitionId}
          onChange={(e) => setPartitionId(e.target.value)}
          placeholder="Ingrese el ID de la partición"
        />
      </div>
      <div className="mb-3">
        <label>Usuario</label>
        <input
          type="text"
          className="form-control"
          value={userId}
          onChange={(e) => setUserId(e.target.value)}
          placeholder="Ingrese su usuario"
        />
      </div>
      <div className="mb-3">
        <label>Contraseña</label>
        <input
          type="password"
          className="form-control"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="Ingrese su contraseña"
        />
      </div>
      <button className="btn btn-primary" onClick={handleLogin}>
        Iniciar sesión
      </button>
      <pre>{output}</pre>
    </div>
  );
};

export default Login;
