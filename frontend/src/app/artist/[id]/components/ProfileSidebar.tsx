"use client";

import { Button } from "@/components/ui/Button";
import ProfileImage from "./ProfileImage";

interface ProfileSidebarProps {
  imageUrl: string;
  isEditing: boolean;
  onEdit: () => void;
  onSave: () => void;
  onCancel: () => void;
  onToggleView: () => void;
  onImageChange?: (newImageUrl: string) => void;
}

export default function ProfileSidebar({ imageUrl, isEditing, onEdit, onSave, onCancel, onToggleView, onImageChange }: ProfileSidebarProps) {

  return (
    <div className="flex flex-col items-center w-full md:w-48">

      <div className="mb-10">
        <ProfileImage
          imageUrl={imageUrl}
          isEditing={isEditing}
          onImageChange={onImageChange}
          className="w-32 h-32 md:w-40 md:h-40 border border-gray-100"
        />
      </div>
      
      
      <div className="flex flex-col w-24 space-y-4">
        {!isEditing ? (
          <>
            <Button 
              onClick={onEdit} 
              variant="light" 
              icon={
                <img 
                  src="/icons/edit.svg" 
                  alt="Edit" 
                  className="w-4 h-4 object-contain" 
                />
              }
              className="w-full flex justify-center items-center gap-1 text-xs whitespace-nowrap rounded-full font-semibold !border"
            >
              Edit
            </Button>
            <Button 
              onClick={onToggleView} 
              variant="dark" 
              icon={
                <div className="shrink-0 flex items-center justify-center min-w-[20px]">
                  <img 
                    src="/icons/preview.svg" 
                    alt="Preview" 
                    className="w-4 h-4 object-contain brightness-0 invert" 
                  />
                </div>
              }
              className="w-full flex justify-center items-center gap-1 text-xs whitespace-nowrap rounded-full font-semibold"
            >
              Preview
            </Button>
          </>
        ) : (
          <>
            <Button 
              onClick={onSave} 
              variant="dark" 
              className="w-full flex justify-center items-center text-xs whitespace-nowrap rounded-full font-semibold"
            >
              Save
            </Button>
            <Button 
              onClick={onCancel} 
              variant="light" 
              className="w-full flex justify-center items-center text-xs whitespace-nowrap rounded-full font-semibold"
            >
              Cancel
            </Button>
          </>
        )}
      </div>
    </div>
  );
}