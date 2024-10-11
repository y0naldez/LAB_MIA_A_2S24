import React from 'react';

const ExecuteButton = ({ onClick }) => {
  return (
    <button className="btn btn-primary mt-3" onClick={onClick}>
      Ejecutar
    </button>
  );
};

export default ExecuteButton;
