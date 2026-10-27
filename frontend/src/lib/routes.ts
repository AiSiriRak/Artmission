export const routes = {
  home: "/",
  login: "/login",
  register: "/register",
  notification: "/notification",
  settings: "/userprofile",

  artist: {
    profile: "/artist-profile",
  },

  order: {
    history: "/orderhistory",
    detail: (id: string) => `/orderhistory/${id}`,
    payment: (id: string) => `/orderhistory/${encodeURIComponent(id)}/payment`,
  },
} as const;
