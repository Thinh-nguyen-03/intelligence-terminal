import type { Metadata } from "next";
import { JetBrains_Mono } from "next/font/google";
import { Header } from "@/components/layout/Header";
import { StatusBar } from "@/components/layout/StatusBar";
import "./globals.css";

const jetbrainsMono = JetBrains_Mono({
  subsets: ["latin"],
  variable: "--font-mono",
});

export const metadata: Metadata = {
  title: "Intelligence Terminal",
  description: "Regime + Positioning Intelligence Terminal",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className={jetbrainsMono.variable}>
      <body className="flex flex-col h-screen overflow-hidden" suppressHydrationWarning>
        <Header />
        <main className="flex-1 overflow-auto p-3">{children}</main>
        <StatusBar />
      </body>
    </html>
  );
}
