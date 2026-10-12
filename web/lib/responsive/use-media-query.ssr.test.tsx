// @vitest-environment node
import { renderToString } from "react-dom/server";
import { expect, it } from "vitest";

import { useIsCoarsePointer, useIsNarrow, useMediaQuery } from "./index";

// Servidor de verdade: sem `window`. O hook não pode tocar em nada do navegador.
function Probe() {
  const narrow = useIsNarrow();
  const coarse = useIsCoarsePointer();
  const wide = useMediaQuery("(min-width: 1024px)");
  return <p>{`estreito=${narrow} toque=${coarse} largo=${wide}`}</p>;
}

it("sem window, renderiza o valor de desktop", () => {
  expect(typeof window).toBe("undefined");
  expect(renderToString(<Probe />)).toBe("<p>estreito=false toque=false largo=true</p>");
});
