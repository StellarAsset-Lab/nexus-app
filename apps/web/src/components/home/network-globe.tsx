"use client";

import { useEffect, useRef } from "react";
import createGlobe, { type Marker, type Arc } from "cobe";

/**
 * A restrained, decorative globe representing global settlement
 * connectivity — the conceptual "financial infrastructure above Stellar"
 * framing, not a rendering of real transaction data. Marker locations are
 * representative financial hubs, not derived from indexed activity; nothing
 * here claims to be live network state (see spec §22/§23 — no fabricated
 * activity).
 */

const SETTLEMENT_HUBS: Marker[] = [
  { location: [40.7128, -74.006], size: 0.05 }, // New York
  { location: [51.5072, -0.1276], size: 0.05 }, // London
  { location: [1.3521, 103.8198], size: 0.05 }, // Singapore
  { location: [50.1109, 8.6821], size: 0.04 }, // Frankfurt
  { location: [35.6762, 139.6503], size: 0.04 }, // Tokyo
  { location: [-23.5505, -46.6333], size: 0.04 }, // São Paulo
  { location: [25.2048, 55.2708], size: 0.04 }, // Dubai
  { location: [6.5244, 3.3792], size: 0.04 }, // Lagos
];

const CONNECTION_ARCS: Arc[] = [
  { from: SETTLEMENT_HUBS[0]!.location, to: SETTLEMENT_HUBS[1]!.location },
  { from: SETTLEMENT_HUBS[1]!.location, to: SETTLEMENT_HUBS[3]!.location },
  { from: SETTLEMENT_HUBS[1]!.location, to: SETTLEMENT_HUBS[2]!.location },
  { from: SETTLEMENT_HUBS[2]!.location, to: SETTLEMENT_HUBS[4]!.location },
  { from: SETTLEMENT_HUBS[0]!.location, to: SETTLEMENT_HUBS[5]!.location },
  { from: SETTLEMENT_HUBS[2]!.location, to: SETTLEMENT_HUBS[6]!.location },
  { from: SETTLEMENT_HUBS[6]!.location, to: SETTLEMENT_HUBS[7]!.location },
];

const CYAN: [number, number, number] = [0.15, 0.55, 0.85];
const NAVY_BASE: [number, number, number] = [0.06, 0.09, 0.16];
const GLOW: [number, number, number] = [0.2, 0.45, 0.9];

export function NetworkGlobe() {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const prefersReducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    let phi = 0;
    let width = 0;

    const onResize = () => {
      if (canvas) {
        width = canvas.offsetWidth;
      }
    };
    window.addEventListener("resize", onResize);
    onResize();

    const globe = createGlobe(canvas, {
      devicePixelRatio: Math.min(window.devicePixelRatio || 1, 2),
      width: width * 2,
      height: width * 2,
      phi: 0,
      theta: 0.3,
      dark: 1,
      diffuse: 1.2,
      mapSamples: 16000,
      mapBrightness: 6,
      baseColor: NAVY_BASE,
      markerColor: CYAN,
      glowColor: GLOW,
      arcColor: CYAN,
      arcWidth: 1.4,
      markers: SETTLEMENT_HUBS,
      arcs: CONNECTION_ARCS,
    });

    let frameId: number | undefined;
    if (!prefersReducedMotion) {
      const animate = () => {
        phi += 0.0025;
        globe.update({ phi, width: width * 2, height: width * 2 });
        frameId = requestAnimationFrame(animate);
      };
      frameId = requestAnimationFrame(animate);
    }

    return () => {
      if (frameId !== undefined) {
        cancelAnimationFrame(frameId);
      }
      globe.destroy();
      window.removeEventListener("resize", onResize);
    };
  }, []);

  return (
    <div className="relative mx-auto aspect-square w-full max-w-xl" aria-hidden="true">
      <canvas ref={canvasRef} className="h-full w-full opacity-90" style={{ contain: "layout paint size" }} />
    </div>
  );
}
