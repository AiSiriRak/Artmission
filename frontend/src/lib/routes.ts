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
    artist: "/orderartist",
    history: "/orderhistory",
    detail: (id: string) => `/orderhistory/${id}`,
  },
} as const;
