"use client"

import { ArtistMiniProfile } from "@/components/feature/artwork/ArtistMiniProfile";
import ArtworkDetailCard from "@/components/feature/artwork/ArtworkDetailCard";
import ArtworkOrderSection from "@/components/feature/artwork/ArtworkOrderSection";
import MainLayout from "@/components/feature/main/MainLayout"
import { Button } from "@/components/ui/Button";

export default function ArtworkDetailPage() {
    return (
        <MainLayout usertype="customer">
        
        <div className="p-20">
            <Button
              variant="light"
              /*onClick={onBack}*/
              className="text-button mb-10 flex items-center gap-2 !border"
            >
              <span>←</span> Back
            </Button>

            <h2 className="text-h2 mb-10 font-bold text-gray-900">Artwork Detail Page</h2>
            {/* Add your artwork detail page content here */}

            <div className="flex items-start gap-10">
              <ArtworkDetailCard />
              <div className="sticky top-28 flex flex-col gap-y-4 gap-x-auto">
                <ArtistMiniProfile />
                <ArtworkOrderSection />
              </div>
            </div>

        </div>
        </MainLayout>
    );
}