import { Navigate, Route, Routes } from "react-router-dom";

import PaginaRecetas from "@/capacidades/cebo/PaginaRecetas";
import PaginaSinResultados from "@/capacidades/cebo/PaginaSinResultados";
import DisposicionCocina from "@/plataforma/caparazon/DisposicionCocina";
import PaginaAcceso from "@/capacidades/accesos/PaginaAcceso";
import PaginaRegistro from "@/capacidades/accesos/PaginaRegistro";
import PaginaSistemas from "@/capacidades/accesos/PaginaSistemas";
import PaginaCambiarPassword from "@/capacidades/identidad/PaginaCambiarPassword";
import PaginaVerificarSegundoFactor from "@/capacidades/identidad/PaginaVerificarSegundoFactor";
import PaginaSolicitarReceta from "@/capacidades/recuperacion/PaginaSolicitarReceta";
import PaginaPrepararReceta from "@/capacidades/recuperacion/PaginaPrepararReceta";

export default function AplicacionEnrutada() {
  return (
    <Routes>
      <Route path="/" element={<PaginaRecetas />} />
      <Route path="/sin-resultados" element={<PaginaSinResultados />} />

      {/* Recuperación de contraseña disfrazada de recetas */}
      <Route path="/solicitar-receta" element={<PaginaSolicitarReceta />} />
      <Route path="/receta/:codigo" element={<PaginaPrepararReceta />} />

      <Route path="/panel" element={<DisposicionCocina />}>
        <Route index element={<Navigate to="acceso" replace />} />
        <Route path="registro" element={<PaginaRegistro />} />
        <Route path="acceso" element={<PaginaAcceso />} />
        <Route path="sistemas" element={<PaginaSistemas />} />
        <Route path="cambiar-password" element={<PaginaCambiarPassword />} />
        <Route path="verificar" element={<PaginaVerificarSegundoFactor />} />
        <Route path="usuarios" element={<Navigate to="/panel/acceso" replace />} />
        <Route path="inventario" element={<Navigate to="/panel/acceso" replace />} />
        <Route path="registrar" element={<Navigate to="/panel/registro" replace />} />
      </Route>

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
