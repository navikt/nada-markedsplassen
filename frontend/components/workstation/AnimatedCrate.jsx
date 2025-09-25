import React, { useState, useEffect } from "react";

const crateSize = 120;

function getCookie(name) {
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) return parts.pop().split(';').shift();
  return null;
}

function setCookie(name, value, months) {
  const d = new Date();
  d.setMonth(d.getMonth() + months);
  document.cookie = `${name}=${value}; expires=${d.toUTCString()}; path=/`;
}

export default function AnimatedCrate({ children }) {
  const [open, setOpen] = useState(false);
  const [showCrate, setShowCrate] = useState(true);

  useEffect(() => {
    if (typeof document !== 'undefined') {
      if (getCookie('sikkerhetsCrateOpened')) {
        setShowCrate(false);
      }
    }
  }, []);

  return (
    <>
      {showCrate && (
        <>
          <div
            className="animated-crate"
            onClick={() => setOpen(true)}
            style={{
              width: crateSize,
              height: crateSize,
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              cursor: "pointer",
              margin: "40px auto",
              position: "relative",
              boxShadow: "0 4px 16px rgba(0,0,0,0.15)",
              borderRadius: 16,
              background: "#c49a6c",
              border: "4px solid #7a5a2f",
              animation: "shake 0.7s infinite",
              userSelect: "none",
            }}
            title="Trykk for å åpne!"
          >
            <span
              style={{
                fontSize: 48,
                fontWeight: "bold",
                color: "#fff",
                textShadow: "2px 2px 0 #7a5a2f, 0 0 8px #fff",
              }}
            >
              ?
            </span>
          </div>
          {open && (
        <div
          className="crate-modal"
          style={{
            position: "fixed",
            top: 0,
            left: 0,
            width: "100vw",
            height: "100vh",
            background: "rgba(0,0,0,0.4)",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            zIndex: 9999,
            }}
          >
            <div
              style={{
                background: "#fff",
                padding: "32px 48px",
                borderRadius: 16,
                boxShadow: "0 4px 32px rgba(0,0,0,0.2)",
                textAlign: "center",
                minWidth: 240,
                maxHeight: "80vh",
                overflowY: "auto",
                width: "90vw",
                borderRadius: 16,
              }}
            >
            <div style={{ marginBottom: 24 }}>
              {children ? children : <span style={{ fontSize: 32 }}>Hello!</span>}
            </div>
                <button
                  style={{
                    padding: "8px 24px",
                    fontSize: 18,
                    borderRadius: 8,
                    border: "none",
                    background: "#c49a6c",
                    color: "#fff",
                    fontWeight: "bold",
                    cursor: "pointer",
                    boxShadow: "0 2px 8px rgba(0,0,0,0.1)",
                  }}
                  onClick={() => {
                    setCookie('sikkerhetsCrateOpened', 'true', 3);
                    setOpen(false);
                    setShowCrate(false);
                  }}
                >
                 Jeg har lest
                </button>
          </div>
        </div>
      )}
        </>
      )}
      <style>{`
        @keyframes shake {
          0% { transform: translate(0px, 0px) rotate(0deg); }
          10% { transform: translate(-2px, 2px) rotate(-2deg); }
          20% { transform: translate(-4px, 0px) rotate(-4deg); }
          30% { transform: translate(4px, 2px) rotate(4deg); }
          40% { transform: translate(2px, -2px) rotate(2deg); }
          50% { transform: translate(-2px, 2px) rotate(-2deg); }
          60% { transform: translate(-4px, 0px) rotate(-4deg); }
          70% { transform: translate(4px, 2px) rotate(4deg); }
          80% { transform: translate(2px, -2px) rotate(2deg); }
          90% { transform: translate(-2px, 2px) rotate(-2deg); }
          100% { transform: translate(0px, 0px) rotate(0deg); }
          }
          @media (max-width: 600px) {
            .crate-modal > div {
              padding: 16px 8px !important;
              min-width: 0 !important;
              width: 98vw !important;
              max-width: 98vw !important;
            }
          }
        `}</style>
    </>
  );
}
