import React from 'react';

const Output = ({ output }) => {
  return (
    <pre className="bg-dark text-light p-3 rounded mt-3" style={{ height: '150px', overflowY: 'auto' }}>
      {output}
    </pre>
  );
};

export default Output;
