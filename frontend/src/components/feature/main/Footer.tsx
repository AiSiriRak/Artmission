import Link from "next/link";

export default function Footer() {
  return (
    <footer className="bg-black text-white py-16 px-8 border-t border-gray-800">
      {/* ใช้ max-w-[1400px] เท่ากับ Header เพื่อความเป๊ะของการจัด Layout */}
      <div className="max-w-[1400px] mx-auto grid grid-cols-1 md:grid-cols-[1.5fr_1fr_1fr] gap-12 md:gap-8">
        
        {/* Column 1: Logo & Description */}
        <div className="flex flex-col items-center md:items-start text-center md:text-left">
          
          {/* Logo Section (ดึงดีไซน์ตาม header.tsx มาใช้) */}
          <Link href="/" className="flex items-center gap-2.5 mb-6">
            {/* 1. รูปไอคอนตัว A (ขาตั้งวาดรูป) */}
            <img 
              src="/icons/A_logo.svg" 
              alt="Artmission Logo" 
              className="w-9 h-10 object-contain" 
            />
            
            {/* 2. ข้อความ RTMISSION ใช้ class สี text-accent-500 เดียวกันกับ Header */}
            <span className="font-extrabold text-accent-500 tracking-wider text-2xl leading-none">
              RTMISSION
            </span>
          </Link>

          <p className="text-gray-300 text-sm leading-relaxed max-w-sm">
            Artmission: The art platform that integrates a creative community by giving indie artists a space to sell their work, while making it easy and accessible for buyers to discover them.
          </p>
        </div>

        {/* Column 2: Members */}
        <div className="flex flex-col items-center md:items-start">
          <h3 className="text-accent-500 text-xl font-bold mb-6">Members</h3>
          <ul className="text-gray-300 text-sm space-y-2 text-center md:text-left">
            <li>Navaporn Homjundee 6730265821</li>
            <li>Napas Ruttanapunyagorn 6730262921</li>
            <li>Kochakorn Srichay 6731301921</li>
            <li>Pumwaree Pipithsukunt 6731338121</li>
            <li>Siripudsorn Raksutakan 6731353521</li>
            <li>Thirada Thomnam 6732015621</li>
            <li>Peeravas Piboolvorakul 6732033921</li>
            <li>Ashira Aungsumal 6732041921</li>
            <li>Akkharaphon Chotwatthakawanit 6631361421</li>
          </ul>
        </div>

        {/* Column 3: Contact */}
        <div className="flex flex-col items-center md:items-start">
          <h3 className="text-accent-500 text-xl font-bold mb-6">Contact</h3>
          <ul className="text-gray-300 text-sm space-y-4">
            {/* Mail */}
            <li className="flex items-center gap-3">
              <svg className="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.5" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"></path>
              </svg>
              <span>[Mail Name]</span>
            </li>
            {/* Facebook */}
            <li className="flex items-center gap-3">
              <svg className="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.5" d="M18 2h-3a5 5 0 00-5 5v3H7v4h3v8h4v-8h3.81l.33-4H14V7a1 1 0 011-1h3V2z"></path>
              </svg>
              <span>[Facebook Name]</span>
            </li>
            {/* Instagram */}
            <li className="flex items-center gap-3">
              <svg className="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                <rect x="2" y="2" width="20" height="20" rx="5" ry="5" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"></rect>
                <path d="M16 11.37A4 4 0 1112.63 8 4 4 0 0116 11.37z" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"></path>
                <line x1="17.5" y1="6.5" x2="17.51" y2="6.5" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"></line>
              </svg>
              <span>[Instagram Name]</span>
            </li>
            {/* Tel */}
            <li className="flex items-center gap-3">
              <svg className="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.5" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z"></path>
              </svg>
              <span>[Tel Number]</span>
            </li>
          </ul>
        </div>

      </div>
    </footer>
  );
}