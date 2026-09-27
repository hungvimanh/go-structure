using Thinktecture;

namespace CHECKIN.Repositories
{
    public interface IMajorRepository
    {
        Task<int> Count(MajorFilter MajorFilter);
        Task<List<Major>> List(MajorFilter MajorFilter);
        Task<Major> Get(long Id);
        Task<bool> Create(Major Major);
        Task<bool> Update(Major Major);
        Task<bool> Delete(Major Major);
    }
    public class MajorRepository : IMajorRepository
    {
        private readonly DataContext DataContext;
        public MajorRepository(DataContext DataContext)
        {
            this.DataContext = DataContext;
        }

        private async Task<IQueryable<MajorDAO>> DynamicFilter(IQueryable<MajorDAO> query, MajorFilter filter)
        {
            if (filter == null)
                return query.Where(x => false);
            query = query.Where(x => !x.DeletedAt.HasValue);
            query = query.Where(x => x.CreatedAt, filter.CreatedAt);
            query = query.Where(x => x.UpdatedAt, filter.UpdatedAt);
            query = query.Where(x => x.Id, filter.Id);
            query = query.Where(x => x.Code, filter.Code);
            query = query.Where(x => x.Name, filter.Name);
            query = query.Where(x => x.OrganizationId, filter.OrganizationId);
            query = query.Where(x => x.StatusId, filter.StatusId);
            if (filter.Search != null)
            {
                query = query.Where(x =>
                   (filter.SearchBy.Contains(MajorSearch.Code) && x.Code.ToLower().Contains(filter.Search.ToLower())) ||
                   (filter.SearchBy.Contains(MajorSearch.Name) && x.Name.ToLower().Contains(filter.Search.ToLower())));
            }

            return query;
        }

        private IQueryable<MajorDAO> DynamicOrder(IQueryable<MajorDAO> query, MajorFilter filter)
        {
            Dictionary<MajorOrder, LambdaExpression> CustomOrder = new Dictionary<MajorOrder, LambdaExpression>()
            {
            };
            query = query.OrderBy(filter.OrderBy, filter.OrderType, CustomOrder);
            query = query.Paging(filter);
            return query;
        }

        private async Task<List<Major>> DynamicSelect(IQueryable<MajorDAO> query, MajorFilter filter)
        {
            List<Major> Majors = await query.Select(x => new Major()
            {
                Id = filter.Selects.Contains(MajorSelect.Id) ? x.Id : default(long),
                Code = filter.Selects.Contains(MajorSelect.Code) ? x.Code : default(string),
                Name = filter.Selects.Contains(MajorSelect.Name) ? x.Name : default(string),
                OrganizationId = filter.Selects.Contains(MajorSelect.Organization) ? x.OrganizationId : default(long),
                StatusId = filter.Selects.Contains(MajorSelect.Status) ? x.StatusId : default(long),
                CreatedAt = x.CreatedAt,
                UpdatedAt = x.UpdatedAt,
                DeletedAt = x.DeletedAt,
                RowId = x.RowId,
            }).ToListAsync();

            Majors.ForEach(x =>
            {
                x.Organization = filter.Selects.Contains(MajorSelect.Organization) ? Organization.Get(x.OrganizationId) : null;
                x.Status = filter.Selects.Contains(MajorSelect.Status) ? Status.Get(x.StatusId) : null;
            });

            return Majors;
        }

        public async Task<int> Count(MajorFilter filter)
        {
            IQueryable<MajorDAO> MajorDAOs = DataContext.Major.AsNoTracking();
            MajorDAOs = await DynamicFilter(MajorDAOs, filter);
            return await MajorDAOs.CountAsync();
        }

        public async Task<List<Major>> List(MajorFilter filter)
        {
            if (filter == null) return new List<Major>();
            IQueryable<MajorDAO> MajorDAOs = DataContext.Major.AsNoTracking();
            MajorDAOs = await DynamicFilter(MajorDAOs, filter);
            MajorDAOs = DynamicOrder(MajorDAOs, filter);
            List<Major> Majors = await DynamicSelect(MajorDAOs, filter);
            return Majors;
        }

        public async Task<Major> Get(long Id)
        {
            Major Major = await DataContext.Major.AsNoTracking()
            .Where(x => x.Id == Id)
            .Where(x => x.DeletedAt == null)
            .Select(x => new Major()
            {
                CreatedAt = x.CreatedAt,
                UpdatedAt = x.UpdatedAt,
                Id = x.Id,
                Code = x.Code,
                Name = x.Name,
                OrganizationId = x.OrganizationId,
                StatusId = x.StatusId,
                RowId = x.RowId,
            }).FirstOrDefaultAsync();

            if (Major == null)
                return null;

            Major.Organization = Organization.Get(Major.OrganizationId);
            Major.Status = Status.Get(Major.StatusId);
            Major.MajorSubjectMappings = await DataContext.MajorSubjectMapping.AsNoTracking()
                .Where(x => x.MajorId == Major.Id)
                .Where(x => x.Subject.DeletedAt == null)
                .Select(x => new MajorSubjectMapping
                {
                    MajorId = x.MajorId,
                    SubjectId = x.SubjectId,
                    Subject = x.Subject == null ? null : new Subject
                    {
                        Id = x.Subject.Id,
                        Code = x.Subject.Code,
                        Name = x.Subject.Name,
                        StatusId = x.Subject.StatusId,
                    },
                }).ToListAsync();

            return Major;
        }
        public async Task<bool> Create(Major Major)
        {
            MajorDAO MajorDAO = new MajorDAO();
            MajorDAO.Id = EntityExtension.SnowflakeId;
            MajorDAO.Code = Major.Code;
            MajorDAO.Name = Major.Name;
            MajorDAO.StatusId = Major.StatusId;
            MajorDAO.OrganizationId = Major.OrganizationId;
            MajorDAO.CreatedAt = StaticParams.DateTimeNow;
            MajorDAO.UpdatedAt = StaticParams.DateTimeNow;
            MajorDAO.RowId = Guid.NewGuid();
            DataContext.Major.Add(MajorDAO);
            await DataContext.SaveChangesAsync();
            Major.Id = MajorDAO.Id;
            await SaveReference(Major);
            return true;
        }

        public async Task<bool> Update(Major Major)
        {
            MajorDAO MajorDAO = DataContext.Major
                .Where(x => x.Id == Major.Id)
                .FirstOrDefault();
            if (MajorDAO == null)
                return false;
            MajorDAO.Code = Major.Code;
            MajorDAO.Name = Major.Name;
            MajorDAO.StatusId = Major.StatusId;
            MajorDAO.OrganizationId = Major.OrganizationId;
            MajorDAO.UpdatedAt = StaticParams.DateTimeNow;
            await DataContext.SaveChangesAsync();
            await SaveReference(Major);
            return true;
        }

        public async Task<bool> Delete(Major Major)
        {
            await DataContext.MajorSubjectMapping.Where(x => x.MajorId == Major.Id).ExecuteDeleteAsync();
            await DataContext.Major
                .Where(x => x.Id == Major.Id)
                .ExecuteUpdateAsync(x => x
                    .SetProperty(y => y.DeletedAt, y => StaticParams.DateTimeNow)
                    .SetProperty(y => y.UpdatedAt, y => StaticParams.DateTimeNow)
                );
            return true;
        }

        private async Task SaveReference(Major Major)
        {
            await DataContext.MajorSubjectMapping
                .Where(x => x.MajorId == Major.Id)
                .ExecuteDeleteAsync();
            if (Major.MajorSubjectMappings?.Any() ?? false)
            {
                foreach (MajorSubjectMapping MajorSubjectMapping in Major.MajorSubjectMappings)
                {
                    MajorSubjectMappingDAO MajorSubjectMappingDAO = new MajorSubjectMappingDAO();
                    MajorSubjectMappingDAO.MajorId = Major.Id;
                    MajorSubjectMappingDAO.SubjectId = MajorSubjectMapping.SubjectId;
                    DataContext.MajorSubjectMapping.Add(MajorSubjectMappingDAO);
                }
                await DataContext.SaveChangesAsync();
            }
        }

    }
}
