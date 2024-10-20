const CommandService = {
  // Función para procesar un comando y enviar al backend
  parseAndSend: async (command) => {
    const parsedCommand = CommandService.parseCommand(command);
    if (parsedCommand) {
      const result = await CommandService.sendCommand(parsedCommand);
      
      // Detectar si el comando es mkdisk y almacenar el path del disco en localStorage
      if (command.startsWith("mkdisk")) {
        const disks = JSON.parse(localStorage.getItem("disks")) || [];
        disks.push(parsedCommand.body.path);  // Almacenar solo el path del disco creado
        localStorage.setItem("disks", JSON.stringify(disks));
      }

      return result;
    } else {
      throw new Error("Comando no válido o no soportado");
    }
  },

  // Función para parsear el comando (como mkdisk, fdisk, etc.)
  parseCommand: (command) => {
    const params = CommandService.parseParams(command);

    if (command.startsWith("mkdisk")) {
      return {
        url: "http://localhost:8080/api/mkdisk",
        method: "POST",
        body: {
          size: parseInt(params.size, 10),
          fit: params.fit.toLowerCase(),
          unit: params.unit.toLowerCase(),
          path: params.path
        }
      };
    } else if (command.startsWith("fdisk")) {
      if (params.delete) {
        return {
          url: "http://localhost:8080/api/fdisk",
          method: "POST",
          body: {
            delete: params.delete.toLowerCase(),  // Fast o Full
            path: params.path,
            name: params.name.toLowerCase()
          }
        };
      } else if (params.add) {
        return {
          url: "http://localhost:8080/api/fdisk",
          method: "POST",
          body: {
            add: parseInt(params.add, 10),
            unit: params.unit.toLowerCase(),
            path: params.path,
            name: params.name.toLowerCase()
          }
        };
      } else {
        return {
          url: "http://localhost:8080/api/fdisk",
          method: "POST",
          body: {
            size: parseInt(params.size, 10),
            path: params.path,
            name: params.name.toLowerCase(),
            unit: params.unit.toLowerCase(),
            type: params.type.toLowerCase(),
            fit: params.fit ? params.fit.toLowerCase() : "w"
          }
        };
      }
    } else if (command.startsWith("mount")) {
      return {
        url: "http://localhost:8080/api/mount",
        method: "POST",
        body: {
          path: params.path,
          name: params.name.toLowerCase()
        }
      };
    } else if (command.startsWith("mkfs")) {
      return {
        url: "http://localhost:8080/api/mkfs",
        method: "POST",
        body: {
          id: params.id,
          type: params.type.toLowerCase(),
          fs: params.fs ? params.fs.toLowerCase() : "2fs"
        }
      };
    } else if (command.startsWith("login")) {
      return {
        url: "http://localhost:8080/api/login",
        method: "POST",
        body: {
          user: params.user,
          pass: params.pass,
          id: params.id
        }
      };
    } else if (command.startsWith("readmbr")) {
      return {
        url: "http://localhost:8080/api/readmbr",
        method: "POST",
        body: {
          path: params.path  
        }
      };
  }else if (command.startsWith("rep")) {
      return {
        url: "http://localhost:8080/api/rep",
        method: "POST",
        body: {
          name: params.name,
          path: params.path,
          id: params.id,
          path_file_ls: params.path_file_ls || ""
        }
      };
    }
    return null;
  },

  // Función para extraer parámetros de los comandos
  parseParams: (command) => {
    const regex = /-(\w+)=("[^"]*"|\S+)/g;
    let match;
    const params = {};
    while ((match = regex.exec(command)) !== null) {
      const key = match[1];
      // Eliminar comillas alrededor de los valores
      const value = match[2].replace(/"/g, ""); 
      params[key] = value;
    }
    return params;
  },

  // Función para enviar los comandos al backend con mejor manejo de errores
  sendCommand: async (parsedCommand) => {
    try {
      const response = await fetch(parsedCommand.url, {
        method: parsedCommand.method,
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify(parsedCommand.body)
      });

      // Si el estado no es ok, arroja un error con el contenido de la respuesta
      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || `Error en el servidor: ${response.statusText}`);
      }

      // Devolver la respuesta en JSON si todo está bien
      return await response.json();
    } catch (error) {
      // Capturar cualquier error durante la solicitud
      throw new Error(error.message || "Error al enviar el comando");
    }
  }
};

export default CommandService;
