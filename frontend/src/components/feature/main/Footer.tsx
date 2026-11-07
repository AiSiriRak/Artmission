import Link from "next/link";
import Image from "next/image";

export default function Footer() {
  return (
    <footer className="bg-black text-white py-12 px-8 border-t border-gray-800">

      <div className="max-w-[1400px] mx-auto grid grid-cols-1 md:grid-cols-[2fr_1fr_1fr] gap-12 md:gap-20">
        {/* Column 1: Logo & Description */}
        <div className="flex flex-col items-center text-center justify-center">

          <Link href="/" className="flex items-center gap-2.5 mb-6">
            <Image
              src="/icons/A_logo.svg"
              alt="Artmission Logo"
              width={52}
              height={52}
              className="object-contain"
            />

            <span className="font-extrabold text-accent-500 tracking-wider text-h1 leading-none">
              RTMISSION
            </span>
          </Link>

          <p className="text-secondary-200 text-small leading-relaxed max-w-md">
            Artmission: The art platform that integrates a creative community by
            giving indie artists a space to sell their work, while making it
            easy and accessible for buyers to discover them.
          </p>
        </div>

        {/* Column 2: Members */}
        <div className="flex flex-col items-center md:items-start">
          <h3 className="text-accent-500 text-h2 mb-6">Members</h3>
          <ul className="text-secondary-200 text-small space-y-2 text-center md:text-left">
            <li>Navaporn Homjundee 6730265821</li>
            <li>Napas Ruttanapunyagorn 6730262921</li>
            <li>Kochakorn Srichay 6731301921</li>
            <li>Pumwaree Pipithsukunt 6731338121</li>
            <li>Siripudsorn Raksutakan 6731353521</li>
            <li>Thirada Thomnam 6732015621</li>
            <li>Peeravas Piboolvorakul 6732033921</li>
            <li>Ashira Aungsumal 6732041921</li>
          </ul>
        </div>

        {/* Column 3: About the project */}
        <div className="flex flex-col items-center md:items-start">
          <h3 className="text-accent-500 text-h2 mb-6">About the Project</h3>
          <p className="text-secondary-200 text-small leading-relaxed text-center md:text-left max-w-md">
            Developed for the 2110423 Team Project (Academic Year 2026) using Agile/SCRUM across 3 sprints to practice full-stack software development.
          </p>
        </div>
      </div>
    </footer>
  );
}
