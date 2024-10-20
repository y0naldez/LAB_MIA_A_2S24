import React, { useState, useEffect } from "react";

const Visualizador = () => {
  const [disks, setDisks] = useState([]);
  const [partitions, setPartitions] = useState([]);
  const [selectedDisk, setSelectedDisk] = useState("");

  useEffect(() => {
    // Leer los discos desde localStorage
    const storedDisks = JSON.parse(localStorage.getItem("disks")) || [];
    setDisks(storedDisks);
  }, []);

  // Función para obtener solo el nombre del archivo del path
  const getDiskName = (path) => {
    return path.split('/').pop();
  };

  // Función para obtener las particiones de un disco
  const fetchPartitions = (diskPath) => {
    // Guardar el disco seleccionado
    setSelectedDisk(getDiskName(diskPath));

    // Hacer la solicitud al backend para obtener las particiones
    fetch("http://localhost:8080/api/readmbr", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ path: diskPath }), 
    })
      .then((response) => response.json())
      .then((data) => {
        setPartitions(data || []); 
      })
      .catch((error) => {
        console.error("Error al obtener particiones:", error);
        setPartitions([]); 
      });
  };

  return (
    <div className="container mt-5">
      <h2>Discos Creados</h2>
      <div className="row">
        {disks.length > 0 ? (
          disks.map((disk, index) => (
            <div key={index} className="col-md-3">
              <div
                className="card mb-3"
                style={{ cursor: "pointer" }}
                // Al hacer clic en el disco, obtener particiones
                onClick={() => fetchPartitions(disk)} 
              >
                <div className="card-body">
                  <h5 className="card-title">Disco: {getDiskName(disk)}</h5>
                </div>
              </div>
            </div>
          ))
        ) : (
          <p>No se han creado discos aún.</p>
        )}
      </div>

      {selectedDisk && (
        <div className="mt-5">
          <h3>Particiones del Disco: {selectedDisk}</h3>
          {partitions && partitions.length > 0 ? (
            <ul className="list-group">
              {partitions.map((partition, index) => (
                <li key={index} className="list-group-item">
                  <strong>Nombre:</strong> {partition.name} | <strong>Tipo:</strong> {partition.type} |{" "}
                  <strong>Tamaño:</strong> {partition.size} | <strong>Inicio:</strong> {partition.start}
                </li>
              ))}
            </ul>
          ) : (
            <p>No se encontraron particiones para este disco.</p>
          )}
        </div>
      )}
    </div>
  );
};

export default Visualizador;
